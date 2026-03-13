package model

import "time"

// Ignore stores one persisted user ignore relation row.
type Ignore struct {
	// ID stores stable ignore row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores owner user identifier.
	UserID uint `gorm:"not null;uniqueIndex:idx_user_ignore_pair,priority:1"`
	// IgnoredUserID stores ignored target user identifier.
	IgnoredUserID uint `gorm:"not null;uniqueIndex:idx_user_ignore_pair,priority:2"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}
