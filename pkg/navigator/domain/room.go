package domain

import "time"

// Room defines one navigable room entry.
type Room struct {
	// ID stores stable room identifier.
	ID int
	// OwnerID stores the room creator identifier.
	OwnerID int
	// OwnerName stores the room creator display name.
	OwnerName string
	// Name stores the room display name.
	Name string
	// Description stores the room description.
	Description string
	// State stores the room access state (open, locked, password).
	State string
	// CategoryID stores the navigator category reference.
	CategoryID int
	// MaxUsers stores the room capacity.
	MaxUsers int
	// CurrentUsers stores the current occupant count.
	CurrentUsers int
	// Score stores the room star rating.
	Score int
	// Tags stores the room searchable tags.
	Tags []string
	// TradeMode stores the trade policy (0 = disabled, 1 = owner, 2 = all).
	TradeMode int
	// CreatedAt stores room creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores room update timestamp.
	UpdatedAt time.Time
}

// Favourite defines one per-user favourite room entry.
type Favourite struct {
	// UserID stores the owning user identifier.
	UserID int
	// RoomID stores the favourite room identifier.
	RoomID int
	// CreatedAt stores when the room was favourited.
	CreatedAt time.Time
}
