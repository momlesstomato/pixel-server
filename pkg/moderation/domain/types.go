package domain

import "time"

// TicketStatus defines the lifecycle state of a support ticket.
type TicketStatus string

const (
	// TicketOpen indicates a newly submitted ticket awaiting review.
	TicketOpen TicketStatus = "open"
	// TicketInProgress indicates a ticket assigned to a moderator.
	TicketInProgress TicketStatus = "in_progress"
	// TicketClosed indicates a resolved ticket.
	TicketClosed TicketStatus = "closed"
	// TicketInvalid indicates a ticket dismissed as invalid.
	TicketInvalid TicketStatus = "invalid"
	// TicketAbusive indicates a ticket flagged as abusive.
	TicketAbusive TicketStatus = "abusive"
)

// Ticket represents a call-for-help support ticket.
type Ticket struct {
	// ID stores the unique ticket identifier.
	ID int64
	// ReporterID stores the user who submitted the ticket.
	ReporterID int
	// ReportedID stores the user being reported.
	ReportedID int
	// RoomID stores the room where the incident occurred.
	RoomID int
	// Category stores the ticket category.
	Category string
	// Message stores the reporter description.
	Message string
	// Status stores the ticket lifecycle state.
	Status TicketStatus
	// AssignedTo stores the moderator handling the ticket.
	AssignedTo int
	// CreatedAt stores the submission timestamp.
	CreatedAt time.Time
	// UpdatedAt stores the last modification timestamp.
	UpdatedAt time.Time
	// ClosedAt stores the resolution timestamp.
	ClosedAt *time.Time
}

// WordFilterEntry represents one word filter rule.
type WordFilterEntry struct {
	// ID stores the unique filter identifier.
	ID int64
	// Pattern stores the text pattern to match.
	Pattern string
	// Replacement stores the substitution text.
	Replacement string
	// IsRegex indicates whether the pattern uses regex.
	IsRegex bool
	// Scope stores "global" or "room".
	Scope string
	// RoomID stores the room identifier when scope is room.
	RoomID int
	// Active indicates whether the filter is enabled.
	Active bool
	// CreatedAt stores the creation timestamp.
	CreatedAt time.Time
}

// Preset represents a moderation action template.
type Preset struct {
	// ID stores the unique preset identifier.
	ID int64
	// Category stores the preset category name.
	Category string
	// Name stores the preset display name.
	Name string
	// ActionType stores the default action type.
	ActionType ActionType
	// DefaultDuration stores the default duration in minutes.
	DefaultDuration int
	// DefaultReason stores the default reason text.
	DefaultReason string
	// Active indicates whether the preset is enabled.
	Active bool
	// CreatedAt stores the creation timestamp.
	CreatedAt time.Time
}

// VisitRecord represents one room visit entry.
type VisitRecord struct {
	// ID stores the unique visit identifier.
	ID int64
	// UserID stores the visiting user.
	UserID int
	// RoomID stores the visited room.
	RoomID int
	// VisitedAt stores the visit timestamp.
	VisitedAt time.Time
}
