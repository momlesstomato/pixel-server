package domain

import "time"

// ActionScope defines the scope of a moderation action.
type ActionScope string

const (
	// ScopeRoom defines a room-scoped moderation action.
	ScopeRoom ActionScope = "room"
	// ScopeHotel defines a hotel-scoped moderation action.
	ScopeHotel ActionScope = "hotel"
)

// ActionType defines the type of moderation action.
type ActionType string

const (
	// TypeKick defines a kick moderation action.
	TypeKick ActionType = "kick"
	// TypeBan defines a ban moderation action.
	TypeBan ActionType = "ban"
	// TypeMute defines a mute moderation action.
	TypeMute ActionType = "mute"
	// TypeWarn defines a warn moderation action.
	TypeWarn ActionType = "warn"
	// TypeTradeLock defines a trade lock moderation action.
	TypeTradeLock ActionType = "trade_lock"
)

// Action represents a single moderation action record.
type Action struct {
	// ID stores the unique action identifier.
	ID int64
	// Scope stores whether this is a room or hotel action.
	Scope ActionScope
	// ActionType stores the kind of moderation action.
	ActionType ActionType
	// TargetUserID stores the user being moderated.
	TargetUserID int
	// IssuerID stores the staff member who issued the action.
	IssuerID int
	// RoomID stores the room identifier when scope is room.
	RoomID int
	// Reason stores the human-readable reason for the action.
	Reason string
	// DurationMinutes stores the intended duration in minutes.
	DurationMinutes int
	// ExpiresAt stores the computed expiry timestamp.
	ExpiresAt *time.Time
	// Active stores whether the action is currently in effect.
	Active bool
	// DeactivatedBy stores which staff member lifted the action.
	DeactivatedBy int
	// DeactivatedAt stores when the action was lifted.
	DeactivatedAt *time.Time
	// IPAddress stores the IP for hotel IP bans.
	IPAddress string
	// MachineID stores the machine fingerprint for hotel machine bans.
	MachineID string
	// CreatedAt stores when the action was created.
	CreatedAt time.Time
}
