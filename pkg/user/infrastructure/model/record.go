package model

import (
	"time"

	"gorm.io/gorm"
)

// Record stores one persisted user row.
type Record struct {
	// ID stores stable user identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Username stores unique user name value.
	Username string `gorm:"size:64;uniqueIndex;not null"`
	// OwnerID stores optional creator owner identifier.
	OwnerID *uint `gorm:"index"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
	// DeletedAt stores row soft-delete timestamp.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
