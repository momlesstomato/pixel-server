package domain

import "context"

// ListFilter defines query filters for listing moderation actions.
type ListFilter struct {
	// Scope filters by action scope.
	Scope ActionScope
	// ActionType filters by action type.
	ActionType ActionType
	// TargetUserID filters by target user.
	TargetUserID int
	// RoomID filters by room identifier.
	RoomID int
	// Active filters by active status.
	Active *bool
	// Offset stores pagination offset.
	Offset int
	// Limit stores page size.
	Limit int
}

// ActionRepository defines persistence behavior for moderation actions.
type ActionRepository interface {
	// Create persists a new moderation action.
	Create(ctx context.Context, action *Action) error
	// FindByID retrieves one action by identifier.
	FindByID(ctx context.Context, id int64) (*Action, error)
	// List returns actions matching the given filter.
	List(ctx context.Context, filter ListFilter) ([]Action, error)
	// Deactivate marks one action as inactive.
	Deactivate(ctx context.Context, id int64, deactivatedBy int) error
	// Delete hard-deletes one room-scoped action.
	Delete(ctx context.Context, id int64) error
	// HasActiveBan checks for active ban on a user by scope.
	HasActiveBan(ctx context.Context, userID int, scope ActionScope) (bool, error)
	// HasActiveMute checks for active mute on a user by scope.
	HasActiveMute(ctx context.Context, userID int, scope ActionScope) (bool, error)
	// HasActiveIPBan checks for active ban on an IP address.
	HasActiveIPBan(ctx context.Context, ip string) (bool, error)
	// HasActiveTradeLock checks for active trade lock on a user.
	HasActiveTradeLock(ctx context.Context, userID int) (bool, error)
}

// TicketRepository defines persistence behavior for support tickets.
type TicketRepository interface {
	// Create persists a new support ticket.
	Create(ctx context.Context, ticket *Ticket) error
	// FindByID retrieves one ticket by identifier.
	FindByID(ctx context.Context, id int64) (*Ticket, error)
	// List returns tickets filtered by status with a limit.
	List(ctx context.Context, status TicketStatus, limit int) ([]Ticket, error)
	// UpdateStatus changes ticket status and optional assignee.
	UpdateStatus(ctx context.Context, id int64, status TicketStatus, assignedTo int) error
	// Delete hard-deletes one ticket row.
	Delete(ctx context.Context, id int64) error
}

// WordFilterRepository defines persistence behavior for word filter rules.
type WordFilterRepository interface {
	// Create persists a new word filter entry.
	Create(ctx context.Context, entry *WordFilterEntry) error
	// FindByID retrieves one word filter entry by identifier.
	FindByID(ctx context.Context, id int64) (*WordFilterEntry, error)
	// ListActive returns active filters for the given scope and room.
	ListActive(ctx context.Context, scope string, roomID int) ([]WordFilterEntry, error)
	// Update persists changes to a word filter entry.
	Update(ctx context.Context, entry *WordFilterEntry) error
	// Delete hard-deletes one word filter entry.
	Delete(ctx context.Context, id int64) error
}

// PresetRepository defines persistence behavior for moderation presets.
type PresetRepository interface {
	// Create persists a new moderation preset.
	Create(ctx context.Context, preset *Preset) error
	// FindByID retrieves one preset by identifier.
	FindByID(ctx context.Context, id int64) (*Preset, error)
	// ListActive returns all active moderation presets.
	ListActive(ctx context.Context) ([]Preset, error)
	// Update persists changes to a moderation preset.
	Update(ctx context.Context, preset *Preset) error
	// Delete hard-deletes one preset.
	Delete(ctx context.Context, id int64) error
}

// VisitRepository defines persistence behavior for room visit records.
type VisitRepository interface {
	// Record persists a new room visit entry.
	Record(ctx context.Context, record *VisitRecord) error
	// ListByUser returns recent visits for one user.
	ListByUser(ctx context.Context, userID int, limit int) ([]VisitRecord, error)
	// ListByRoom returns recent visits for one room.
	ListByRoom(ctx context.Context, roomID int, limit int) ([]VisitRecord, error)
}
