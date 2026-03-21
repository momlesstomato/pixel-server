package model

import "time"

// Item stores one owned furniture instance row in PostgreSQL.
type Item struct {
	// ID stores stable item instance identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the item owner identifier.
	UserID uint `gorm:"not null;index"`
	// RoomID stores the placed room identifier, zero when in inventory.
	RoomID uint `gorm:"not null;default:0;index"`
	// DefinitionID stores the item definition foreign key.
	DefinitionID uint `gorm:"not null;index"`
	// ExtraData stores item-specific custom data payload.
	ExtraData string `gorm:"type:text;not null;default:''"`
	// LimitedNumber stores the limited edition serial number.
	LimitedNumber int `gorm:"not null;default:0"`
	// LimitedTotal stores the limited edition total print run.
	LimitedTotal int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Item.
func (Item) TableName() string { return "items" }
