package domain

import "time"

// Item defines one owned furniture instance.
type Item struct {
	// ID stores stable item instance identifier.
	ID int
	// UserID stores the item owner identifier.
	UserID int
	// RoomID stores the placed room identifier, zero when in inventory.
	RoomID int
	// DefinitionID stores the item definition foreign key.
	DefinitionID int
	// ExtraData stores item-specific custom data payload.
	ExtraData string
	// LimitedNumber stores the limited edition serial number.
	LimitedNumber int
	// LimitedTotal stores the limited edition total print run.
	LimitedTotal int
	// CreatedAt stores item creation timestamp.
	CreatedAt time.Time
}
