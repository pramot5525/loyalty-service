package repository

import (
	"context"

	"loyalty-service/internal/core/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Upsert(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "external_id"}},
			DoNothing: true,
		}).
		Create(user).Error
}

func (r *userRepository) UpdatePoints(ctx context.Context, userID uint, balanceDelta int, pendingDelta int) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"point_balance": gorm.Expr("point_balance + ?", balanceDelta),
			"pending_point": gorm.Expr("pending_point + ?", pendingDelta),
		}).Error
}
