package services

import (
	"context"
	"fmt"
	"math"

	"loyalty-service/internal/core/domain"
	"loyalty-service/internal/core/ports/input"
	"loyalty-service/internal/core/ports/output"
)

type OrderService struct {
	userRepo    output.UserRepository
	orderRepo   output.OrderRepository
	pointTxRepo output.PointTransactionRepository
	transactor  output.Transactor
}

func NewOrderService(
	userRepo output.UserRepository,
	orderRepo output.OrderRepository,
	pointTxRepo output.PointTransactionRepository,
	transactor output.Transactor,
) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		pointTxRepo: pointTxRepo,
		transactor:  transactor,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, in input.CreateOrderInput) (*domain.Order, error) {
	// 2. Compute points
	netPrice := in.TotalFromBuyer
	points := int(math.Trunc(netPrice / pointRate))

	var order *domain.Order
	err := s.transactor.WithTx(ctx, func(ctx context.Context) error {
		// 1. Upsert user
		if err := s.userRepo.Upsert(ctx, &domain.User{ExternalID: in.ExternalUserID}); err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}
		user, err := s.userRepo.FindByExternalID(ctx, in.ExternalUserID)
		if err != nil {
			return fmt.Errorf("find user: %w", err)
		}

		// 3. Create order
		order = &domain.Order{
			ExternalOrderID: in.ExternalOrderID,
			UserID:          user.ID,
			TotalFromBuyer:  in.TotalFromBuyer,
			NetPrice:        netPrice,
			EarnedPoint:     points,
			Status:          domain.OrderStatusPending,
		}
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		// 4. Create point transaction (PENDING)
		pointTx := &domain.PointTransaction{
			UserID:  user.ID,
			OrderID: order.ID,
			Point:   points,
			Type:    domain.TransactionTypeEarn,
			Status:  domain.TransactionStatusPending,
		}
		if err := s.pointTxRepo.Create(ctx, pointTx); err != nil {
			return fmt.Errorf("create point transaction: %w", err)
		}

		// 5. Increment user pending_point
		if err := s.userRepo.UpdatePoints(ctx, user.ID, 0, points); err != nil {
			return fmt.Errorf("update pending points: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) CompleteOrder(ctx context.Context, externalOrderID string) error {
	order, err := s.orderRepo.FindByExternalID(ctx, externalOrderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != domain.OrderStatusPending {
		return fmt.Errorf("order is not in PENDING status")
	}

	tx, err := s.pointTxRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("point transaction not found: %w", err)
	}

	if err := s.pointTxRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCompleted); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}

	// Move points: pending → balance
	if err := s.userRepo.UpdatePoints(ctx, order.UserID, order.EarnedPoint, -order.EarnedPoint); err != nil {
		return fmt.Errorf("update user points: %w", err)
	}

	return s.orderRepo.UpdateStatus(ctx, order.ID, domain.OrderStatusCompleted)
}

func (s *OrderService) CancelOrder(ctx context.Context, externalOrderID string) error {
	order, err := s.orderRepo.FindByExternalID(ctx, externalOrderID)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if order.Status != domain.OrderStatusPending {
		return fmt.Errorf("order is not in PENDING status")
	}

	tx, err := s.pointTxRepo.FindByOrderID(ctx, order.ID)
	if err != nil {
		return fmt.Errorf("point transaction not found: %w", err)
	}

	if err := s.pointTxRepo.UpdateStatus(ctx, tx.ID, domain.TransactionStatusCancelled); err != nil {
		return fmt.Errorf("update transaction status: %w", err)
	}

	// Revert pending points
	if err := s.userRepo.UpdatePoints(ctx, order.UserID, 0, -order.EarnedPoint); err != nil {
		return fmt.Errorf("revert pending points: %w", err)
	}

	return s.orderRepo.UpdateStatus(ctx, order.ID, domain.OrderStatusCancelled)
}
