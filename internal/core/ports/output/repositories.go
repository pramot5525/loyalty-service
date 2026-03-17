package output

import (
	"context"
	"loyalty-service/internal/core/domain"
)

type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepository interface {
	FindByExternalID(ctx context.Context, externalID string) (*domain.User, error)
	Upsert(ctx context.Context, user *domain.User) error
	UpdatePoints(ctx context.Context, userID uint, balanceDelta int, pendingDelta int) error
}

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	FindByExternalID(ctx context.Context, externalOrderID string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uint, status domain.OrderStatus) error
}

type PointTransactionRepository interface {
	Create(ctx context.Context, tx *domain.PointTransaction) error
	FindByOrderID(ctx context.Context, orderID uint) (*domain.PointTransaction, error)
	UpdateStatus(ctx context.Context, id uint, status domain.TransactionStatus) error
}
