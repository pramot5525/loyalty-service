package http

import (
	"loyalty-service/internal/adapters/http/handler"
	"loyalty-service/internal/core/ports/input"

	"github.com/gofiber/fiber/v2"
)

func NewRouter(orderUC input.OrderUseCase, pointUC input.PointUseCase) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	orderH := handler.NewOrderHandler(orderUC)
	pointH := handler.NewPointHandler(pointUC)

	app.Get("/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	v1 := app.Group("/v1")

	v1.Post("/orders", orderH.CreateOrder)
	v1.Put("/orders/:id/complete", orderH.CompleteOrder)
	v1.Put("/orders/:id/cancel", orderH.CancelOrder)

	v1.Post("/points/calculate", pointH.CalculatePoint)
	v1.Get("/users/:id/points", pointH.GetUserPoints)

	return app
}
