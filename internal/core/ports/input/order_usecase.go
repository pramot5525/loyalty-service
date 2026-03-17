package input

import (
	"context"
	"loyalty-service/internal/core/domain"
)

type CreateOrderInput struct {
	ExternalOrderID string
	ExternalUserID  string
	TotalFromBuyer  float64
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error)
	CompleteOrder(ctx context.Context, externalOrderID string) error
	CancelOrder(ctx context.Context, externalOrderID string) error
}
