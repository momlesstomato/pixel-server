package model

import "time"

// LoginEvent stores one successful login event record.
type LoginEvent struct {
	// ID stores stable login event identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores authenticated user identifier.
	UserID int `gorm:"index;not null"`
	// Holder stores application holder identifier.
	Holder string `gorm:"size:120;not null"`
	// LoggedAt stores successful login timestamp in UTC.
	LoggedAt time.Time `gorm:"index;not null"`
}
