package handler

import (
	"encoding/json"

	"loyalty-service/internal/adapters/kafka"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	producer *kafka.Producer
}

func NewOrderHandler(producer *kafka.Producer) *OrderHandler {
	return &OrderHandler{producer: producer}
}

type createOrderItem struct {
	ExternalOrderID string  `json:"external_order_id"`
	ExternalUserID  string  `json:"external_user_id"`
	TotalFromBuyer  float64 `json:"total_from_buyer"`
}

// CreateOrders accepts an array of orders and publishes each as an
// "order.created" event to Kafka. The consumer processes them asynchronously.
func (h *OrderHandler) CreateOrders(c *fiber.Ctx) error {
	var items []createOrderItem
	if err := json.Unmarshal(c.Body(), &items); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if len(items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request body must be a non-empty array"})
	}

	events := make([]kafka.OrderEventMessage, 0, len(items))
	invalid := []string{}

	for _, item := range items {
		if item.ExternalOrderID == "" || item.ExternalUserID == "" {
			invalid = append(invalid, item.ExternalOrderID)
			continue
		}
		events = append(events, kafka.OrderEventMessage{
			Type: "order.created",
			Data: kafka.OrderEventData{
				ExternalOrderID: item.ExternalOrderID,
				ExternalUserID:  item.ExternalUserID,
				TotalFromBuyer:  item.TotalFromBuyer,
			},
		})
	}

	if len(invalid) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "some orders are missing required fields",
			"invalid": invalid,
		})
	}

	if err := h.producer.PublishOrderEvents(c.Context(), events); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"published": len(events),
	})
}

func (h *OrderHandler) CompleteOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.producer.PublishOrderEvents(c.Context(), []kafka.OrderEventMessage{
		{Type: "order.delivered", Data: kafka.OrderEventData{ExternalOrderID: id}},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"published": 1})
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.producer.PublishOrderEvents(c.Context(), []kafka.OrderEventMessage{
		{Type: "order.cancelled", Data: kafka.OrderEventData{ExternalOrderID: id}},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"published": 1})
}
