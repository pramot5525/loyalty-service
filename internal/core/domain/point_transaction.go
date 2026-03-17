package domain

import "time"

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeEarn   TransactionType = "EARN"
	TransactionTypeBurn   TransactionType = "BURN"
	TransactionTypeExpire TransactionType = "EXPIRE"
	TransactionTypeRevert TransactionType = "REVERT"

	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

type PointTransaction struct {
	ID        uint              `gorm:"primaryKey;autoIncrement"`
	UserID    uint              `gorm:"not null;index"`
	OrderID   uint              `gorm:"not null;index"`
	Point     int               `gorm:"not null"`
	Type      TransactionType   `gorm:"type:varchar(20);not null"`
	Status    TransactionStatus `gorm:"type:varchar(20);default:'PENDING';not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
