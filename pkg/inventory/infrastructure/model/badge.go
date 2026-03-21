package model

import "time"

// Badge stores one user badge row in PostgreSQL.
type Badge struct {
	// ID stores stable badge row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the badge owner identifier.
	UserID uint `gorm:"not null;index"`
	// BadgeCode stores the badge type code.
	BadgeCode string `gorm:"size:50;not null"`
	// SlotID stores the equipped badge slot, zero when unequipped.
	SlotID int16 `gorm:"not null;default:0"`
	// CreatedAt stores badge award timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Badge.
func (Badge) TableName() string { return "user_badges" }
