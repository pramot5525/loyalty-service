package input

import "context"

type PointUseCase interface {
	CalculatePoint(netPrice float64) int
	GetUserPointBalance(ctx context.Context, externalUserID string) (balance int, pending int, err error)
}
