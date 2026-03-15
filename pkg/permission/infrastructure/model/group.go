package model

import "time"

// Group defines one persisted permission_groups row.
type Group struct {
	// ID stores stable group identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Name stores unique machine-friendly group name.
	Name string `gorm:"size:64;uniqueIndex;not null"`
	// DisplayName stores human-friendly group label.
	DisplayName string `gorm:"size:128;not null;default:''"`
	// Priority stores group priority for effective-group resolution.
	Priority int `gorm:"not null;default:0"`
	// ClubLevel stores protocol club level attribute.
	ClubLevel int `gorm:"not null;default:0"`
	// SecurityLevel stores protocol security level attribute.
	SecurityLevel int `gorm:"not null;default:0"`
	// IsAmbassador stores protocol ambassador attribute.
	IsAmbassador bool `gorm:"not null;default:false"`
	// IsDefault stores default-assignment marker.
	IsDefault bool `gorm:"not null;default:false;index"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}
