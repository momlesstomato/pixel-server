package model

import "time"

// Room stores one room row in PostgreSQL.
type Room struct {
	// ID stores stable room identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// OwnerID stores the room creator identifier.
	OwnerID uint `gorm:"not null;index:idx_rooms_owner"`
	// OwnerName stores the room creator display name.
	OwnerName string `gorm:"size:64;not null"`
	// Name stores the room display name.
	Name string `gorm:"size:100;not null"`
	// Description stores the room description.
	Description string `gorm:"size:255;not null;default:''"`
	// State stores the room access state.
	State string `gorm:"size:20;not null;default:open"`
	// CategoryID stores the navigator category reference.
	CategoryID uint `gorm:"not null;default:0;index:idx_rooms_category"`
	// MaxUsers stores the room capacity.
	MaxUsers int `gorm:"not null;default:25"`
	// Score stores the room star rating.
	Score int `gorm:"not null;default:0"`
	// Tags stores comma-separated room tags.
	Tags string `gorm:"size:255;not null;default:''"`
	// TradeMode stores the trade policy code.
	TradeMode int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Room.
func (Room) TableName() string { return "rooms" }

// Favourite stores one user-room favourite entry in PostgreSQL.
type Favourite struct {
	// ID stores stable favourite identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the owning user identifier.
	UserID uint `gorm:"not null;uniqueIndex:idx_nav_fav_pair,priority:1"`
	// RoomID stores the favourite room identifier.
	RoomID uint `gorm:"not null;uniqueIndex:idx_nav_fav_pair,priority:2"`
	// CreatedAt stores when the room was favourited.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Favourite.
func (Favourite) TableName() string { return "navigator_favourites" }
