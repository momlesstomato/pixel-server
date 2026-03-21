package domain

import "time"

// MaxBadgeSlots defines the maximum number of visible badge slots.
const MaxBadgeSlots = 5

// Badge defines one user-owned badge entry.
type Badge struct {
	// ID stores stable badge row identifier.
	ID int
	// UserID stores the badge owner identifier.
	UserID int
	// BadgeCode stores the badge type code.
	BadgeCode string
	// SlotID stores the equipped badge slot, zero when unequipped.
	SlotID int
	// CreatedAt stores badge award timestamp.
	CreatedAt time.Time
}

// BadgeSlot maps one equipped badge slot for client display.
type BadgeSlot struct {
	// SlotID stores the slot position from one through five.
	SlotID int
	// BadgeCode stores the badge type code in that slot.
	BadgeCode string
}
