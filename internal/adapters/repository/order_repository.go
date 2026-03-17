package repository

import (
	"context"

	"loyalty-service/internal/core/domain"

	"gorm.io/gorm"
)

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) FindByExternalID(ctx context.Context, externalOrderID string) (*domain.Order, error) {
	var order domain.Order
	err := r.db.WithContext(ctx).Where("external_order_id = ?", externalOrderID).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status domain.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Order{}).
		Where("id = ?", id).
		Update("status", status).Error
}
