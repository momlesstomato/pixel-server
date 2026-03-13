package domain

import "time"

// WardrobeSlot defines one saved outfit slot for one user.
type WardrobeSlot struct {
	// SlotID stores stable slot index value.
	SlotID int
	// Figure stores avatar figure value for the slot.
	Figure string
	// Gender stores avatar gender value for the slot.
	Gender string
}

// RespectRecord defines one stored respect audit event.
type RespectRecord struct {
	// ID stores stable respect record identifier.
	ID int
	// ActorUserID stores user identifier that sent respect.
	ActorUserID int
	// TargetID stores user identifier that received respect.
	TargetID int
	// TargetType stores respect target type marker.
	TargetType RespectTargetType
	// RespectedAt stores event UTC date marker.
	RespectedAt time.Time
}
