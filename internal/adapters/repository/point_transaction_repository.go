package repository

import (
	"context"

	"loyalty-service/internal/core/domain"

	"gorm.io/gorm"
)

type pointTransactionRepository struct {
	db *gorm.DB
}

func NewPointTransactionRepository(db *gorm.DB) *pointTransactionRepository {
	return &pointTransactionRepository{db: db}
}

func (r *pointTransactionRepository) Create(ctx context.Context, tx *domain.PointTransaction) error {
	return dbFromCtx(ctx, r.db).WithContext(ctx).Create(tx).Error
}

func (r *pointTransactionRepository) FindByOrderID(ctx context.Context, orderID uint) (*domain.PointTransaction, error) {
	var tx domain.PointTransaction
	err := dbFromCtx(ctx, r.db).WithContext(ctx).Where("order_id = ?", orderID).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *pointTransactionRepository) UpdateStatus(ctx context.Context, id uint, status domain.TransactionStatus) error {
	return dbFromCtx(ctx, r.db).WithContext(ctx).Model(&domain.PointTransaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}
