package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

func NewProducer(cfg ProducerConfig) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	return &Producer{writer: w}
}

func (p *Producer) PublishOrderEvents(ctx context.Context, events []OrderEventMessage) error {
	msgs := make([]kafka.Message, len(events))
	for i, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal event[%d]: %w", i, err)
		}
		msgs[i] = kafka.Message{
			Key:   []byte(e.Data.ExternalOrderID),
			Value: b,
		}
	}
	return p.writer.WriteMessages(ctx, msgs...)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
