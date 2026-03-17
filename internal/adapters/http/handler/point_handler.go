package handler

import (
	"loyalty-service/internal/core/ports/input"

	"github.com/gofiber/fiber/v2"
)

type PointHandler struct {
	pointUseCase input.PointUseCase
}

func NewPointHandler(uc input.PointUseCase) *PointHandler {
	return &PointHandler{pointUseCase: uc}
}

func (h *PointHandler) CalculatePoint(c *fiber.Ctx) error {
	var body struct {
		NetPrice float64 `json:"net_price"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	point := h.pointUseCase.CalculatePoint(body.NetPrice)
	return c.JSON(fiber.Map{
		"net_price": body.NetPrice,
		"point":     point,
	})
}

func (h *PointHandler) GetUserPoints(c *fiber.Ctx) error {
	userID := c.Params("id")
	balance, pending, err := h.pointUseCase.GetUserPointBalance(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"user_id":       userID,
		"point_balance": balance,
		"pending_point": pending,
	})
}
