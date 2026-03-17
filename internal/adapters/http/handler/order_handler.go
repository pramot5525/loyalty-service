package handler

import (
	"loyalty-service/internal/core/ports/input"

	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderUseCase input.OrderUseCase
}

func NewOrderHandler(uc input.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: uc}
}

type createOrderRequest struct {
	ExternalOrderID          string  `json:"external_order_id"`
	ExternalUserID           string  `json:"external_user_id"`
	TotalFromBuyer           float64 `json:"total_from_buyer"`
	ShippingCost             float64 `json:"shipping_cost"`
	ShippingDiscountBySeller float64 `json:"shipping_discount_by_seller"`
	ShippingDiscountBySystem float64 `json:"shipping_discount_by_system"`
}

func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	var req createOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.ExternalOrderID == "" || req.ExternalUserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "external_order_id and external_user_id are required"})
	}

	order, err := h.orderUseCase.CreateOrder(c.Context(), input.CreateOrderInput{
		ExternalOrderID:          req.ExternalOrderID,
		ExternalUserID:           req.ExternalUserID,
		TotalFromBuyer:           req.TotalFromBuyer,
		ShippingCost:             req.ShippingCost,
		ShippingDiscountBySeller: req.ShippingDiscountBySeller,
		ShippingDiscountBySystem: req.ShippingDiscountBySystem,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":               order.ID,
		"external_order_id": order.ExternalOrderID,
		"net_price":        order.NetPrice,
		"earned_point":     order.EarnedPoint,
		"status":           order.Status,
	})
}

func (h *OrderHandler) CompleteOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.orderUseCase.CompleteOrder(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "order completed"})
}

func (h *OrderHandler) CancelOrder(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.orderUseCase.CancelOrder(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "order cancelled"})
}
