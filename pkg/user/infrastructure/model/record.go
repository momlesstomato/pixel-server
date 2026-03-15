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
	// Figure stores avatar figure string.
	Figure string `gorm:"size:255;not null;default:hr-115-42.hd-180-1.ch-3030-82.lg-275-82.sh-295-62"`
	// Gender stores avatar gender marker.
	Gender string `gorm:"size:1;not null;default:M"`
	// Motto stores user profile motto.
	Motto string `gorm:"size:127;not null;default:''"`
	// RealName stores user display real name.
	RealName string `gorm:"size:64;not null;default:''"`
	// RespectsReceived stores total received user respects.
	RespectsReceived int `gorm:"not null;default:0"`
	// HomeRoomID stores user configured home room identifier.
	HomeRoomID int `gorm:"not null;default:-1"`
	// CanChangeName stores whether user can rename account.
	CanChangeName bool `gorm:"not null;default:false"`
	// NoobnessLevel stores account age tier.
	NoobnessLevel int `gorm:"not null;default:2"`
	// SafetyLocked stores account safety lock state.
	SafetyLocked bool `gorm:"not null;default:false"`
	// LastAccessAt stores last successful access timestamp.
	LastAccessAt *time.Time
	// GroupID stores permission group identifier.
	GroupID uint `gorm:"not null;default:1;index"`
	// OwnerID stores optional creator owner identifier.
	OwnerID *uint `gorm:"index"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
	// DeletedAt stores row soft-delete timestamp.
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the persisted table name for user records.
func (Record) TableName() string {
	return "users"
}
