package domain

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	ExternalID   string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PointBalance int       `gorm:"default:0;not null"`
	PendingPoint int       `gorm:"default:0;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
