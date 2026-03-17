package http

import (
	"loyalty-service/internal/adapters/http/handler"
	"loyalty-service/internal/adapters/kafka"
	"loyalty-service/internal/core/ports/input"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewRouter(pointUC input.PointUseCase, producer *kafka.Producer, corsOrigins string) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
		ExposeHeaders: "Content-Length",
		MaxAge:        300,
	}))

	orderH := handler.NewOrderHandler(producer)
	pointH := handler.NewPointHandler(pointUC)

	app.Get("/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	v1 := app.Group("/v1")

	v1.Post("/orders", orderH.CreateOrders)
	v1.Put("/orders/:id/complete", orderH.CompleteOrder)
	v1.Put("/orders/:id/cancel", orderH.CancelOrder)

	v1.Post("/points/calculate", pointH.CalculatePoint)
	v1.Get("/users/:id/points", pointH.GetUserPoints)

	return app
}
