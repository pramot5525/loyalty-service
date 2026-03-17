package kafka

import (
	"context"
	"encoding/json"
	"log"

	"loyalty-service/internal/core/ports/input"

	"github.com/segmentio/kafka-go"
)

// OrderEventMessage mirrors the Kafka message envelope from the order service.
//
// Example payload:
//
//	{
//	  "type": "order.created",
//	  "data": {
//	    "external_order_id": "ORD-001",
//	    "external_user_id": "USR-001",
//	    "total_from_buyer": 500
//	  }
//	}
//
// Supported types:
//   - order.created   → CreateOrder (pend points)
//   - order.delivered → CompleteOrder (confirm points)
//   - order.cancelled → CancelOrder (revert points)
type OrderEventMessage struct {
	Type string          `json:"type"`
	Data OrderEventData  `json:"data"`
}

type OrderEventData struct {
	ExternalOrderID string  `json:"external_order_id"`
	ExternalUserID  string  `json:"external_user_id"`
	TotalFromBuyer  float64 `json:"total_from_buyer"`
}

type Consumer struct {
	reader       *kafka.Reader
	orderUseCase input.OrderUseCase
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(cfg ConsumerConfig, orderUC input.OrderUseCase) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 1,
		MaxBytes: 10e6, // 10 MB
		Logger:   kafka.LoggerFunc(func(msg string, args ...interface{}) { log.Printf("[kafka] "+msg, args...) }),
	})

	return &Consumer{reader: r, orderUseCase: orderUC}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Printf("[kafka] consumer started — topic: %s", c.reader.Config().Topic)

	go func() {
		defer c.reader.Close()
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("[kafka] consumer stopped")
					return
				}
				log.Printf("[kafka] read error: %v", err)
				continue
			}
			c.handle(ctx, msg)
		}
	}()
}

func (c *Consumer) handle(ctx context.Context, msg kafka.Message) {
	var event OrderEventMessage
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("[kafka] failed to unmarshal message: %v — raw: %s", err, string(msg.Value))
		return
	}

	log.Printf("[kafka] received event type=%s order=%s", event.Type, event.Data.ExternalOrderID)

	var err error
	switch event.Type {
	case "order.created":
		_, err = c.orderUseCase.CreateOrder(ctx, input.CreateOrderInput{
			ExternalOrderID: event.Data.ExternalOrderID,
			ExternalUserID:  event.Data.ExternalUserID,
			TotalFromBuyer:  event.Data.TotalFromBuyer,
		})
	case "order.delivered":
		err = c.orderUseCase.CompleteOrder(ctx, event.Data.ExternalOrderID)
	case "order.cancelled":
		err = c.orderUseCase.CancelOrder(ctx, event.Data.ExternalOrderID)
	default:
		log.Printf("[kafka] ignored unknown event type: %s", event.Type)
		return
	}

	if err != nil {
		log.Printf("[kafka] error processing event type=%s order=%s: %v", event.Type, event.Data.ExternalOrderID, err)
	}
}
