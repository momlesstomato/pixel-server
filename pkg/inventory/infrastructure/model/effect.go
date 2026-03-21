package model

import "time"

// Effect stores one user avatar effect row in PostgreSQL.
type Effect struct {
	// ID stores stable effect row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the effect owner identifier.
	UserID uint `gorm:"not null;index"`
	// EffectID stores the avatar effect type identifier.
	EffectID int `gorm:"not null"`
	// Duration stores total duration in seconds.
	Duration int `gorm:"not null;default:86400"`
	// Quantity stores remaining activations.
	Quantity int `gorm:"not null;default:1"`
	// ActivatedAt stores first activation timestamp.
	ActivatedAt *time.Time
	// IsPermanent stores whether the effect never expires.
	IsPermanent bool `gorm:"not null;default:false"`
	// CreatedAt stores effect award timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Effect.
func (Effect) TableName() string { return "user_effects" }
