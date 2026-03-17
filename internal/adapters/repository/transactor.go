package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func dbFromCtx(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return fallback
}

type transactor struct {
	db *gorm.DB
}

func NewTransactor(db *gorm.DB) *transactor {
	return &transactor{db: db}
}

func (t *transactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}
