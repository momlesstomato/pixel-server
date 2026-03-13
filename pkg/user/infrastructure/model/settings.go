package model

import "time"

// Settings stores one persisted user settings row.
type Settings struct {
	// ID stores stable settings row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores owning user identifier.
	UserID uint `gorm:"not null;uniqueIndex"`
	// VolumeSystem stores global system volume percentage.
	VolumeSystem int `gorm:"not null;default:100"`
	// VolumeFurni stores furniture volume percentage.
	VolumeFurni int `gorm:"not null;default:100"`
	// VolumeTrax stores trax volume percentage.
	VolumeTrax int `gorm:"not null;default:100"`
	// OldChat stores classic chat style preference.
	OldChat bool `gorm:"not null;default:false"`
	// RoomInvites stores room invite preference.
	RoomInvites bool `gorm:"not null;default:true"`
	// CameraFollow stores camera follow mode preference.
	CameraFollow bool `gorm:"not null;default:true"`
	// Flags stores bitmask settings field.
	Flags int `gorm:"not null;default:0"`
	// ChatType stores client chat rendering type.
	ChatType int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}
