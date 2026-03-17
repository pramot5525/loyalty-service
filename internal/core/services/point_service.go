package services

import (
	"context"
	"fmt"
	"math"

	"loyalty-service/internal/core/ports/output"
)

const pointRate = 50

type PointService struct {
	userRepo output.UserRepository
}

func NewPointService(userRepo output.UserRepository) *PointService {
	return &PointService{userRepo: userRepo}
}

func (s *PointService) CalculatePoint(netPrice float64) int {
	if netPrice <= 0 {
		return 0
	}
	return int(math.Trunc(netPrice / pointRate))
}

func (s *PointService) GetUserPointBalance(ctx context.Context, externalUserID string) (int, int, error) {
	user, err := s.userRepo.FindByExternalID(ctx, externalUserID)
	if err != nil {
		return 0, 0, fmt.Errorf("user not found: %w", err)
	}
	return user.PointBalance, user.PendingPoint, nil
}
