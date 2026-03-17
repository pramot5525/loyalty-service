package domain

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusCompleted OrderStatus = "COMPLETED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID                       uint        `gorm:"primaryKey;autoIncrement"`
	ExternalOrderID          string      `gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID                   uint        `gorm:"not null;index"`
	TotalFromBuyer           float64     `gorm:"not null"`
	ShippingCost             float64     `gorm:"default:0"`
	ShippingDiscountBySeller float64     `gorm:"default:0"`
	ShippingDiscountBySystem float64     `gorm:"default:0"`
	NetPrice                 float64     `gorm:"not null"`
	EarnedPoint              int         `gorm:"default:0;not null"`
	Status                   OrderStatus `gorm:"type:varchar(20);default:'PENDING';not null"`
	CreatedAt                time.Time
	UpdatedAt                time.Time
}
