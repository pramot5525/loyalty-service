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
}

func NewOrderService(
	userRepo output.UserRepository,
	orderRepo output.OrderRepository,
	pointTxRepo output.PointTransactionRepository,
) *OrderService {
	return &OrderService{
		userRepo:    userRepo,
		orderRepo:   orderRepo,
		pointTxRepo: pointTxRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, in input.CreateOrderInput) (*domain.Order, error) {
	// 1. Upsert user
	user := &domain.User{ExternalID: in.ExternalUserID}
	if err := s.userRepo.Upsert(ctx, user); err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	user, err := s.userRepo.FindByExternalID(ctx, in.ExternalUserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	// 2. Compute points
	netPrice := in.TotalFromBuyer
	points := int(math.Trunc(netPrice / pointRate))

	// 3. Create order
	order := &domain.Order{
		ExternalOrderID: in.ExternalOrderID,
		UserID:          user.ID,
		TotalFromBuyer:  in.TotalFromBuyer,
		NetPrice:        netPrice,
		EarnedPoint:     points,
		Status:          domain.OrderStatusPending,
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// 4. Create point transaction (PENDING)
	tx := &domain.PointTransaction{
		UserID:  user.ID,
		OrderID: order.ID,
		Point:   points,
		Type:    domain.TransactionTypeEarn,
		Status:  domain.TransactionStatusPending,
	}
	if err := s.pointTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("create point transaction: %w", err)
	}

	// 5. Increment user pending_point
	if err := s.userRepo.UpdatePoints(ctx, user.ID, 0, points); err != nil {
		return nil, fmt.Errorf("update pending points: %w", err)
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
