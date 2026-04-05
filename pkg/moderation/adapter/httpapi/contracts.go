package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// ModerationService defines moderation behavior for HTTP routes.
type ModerationService interface {
	// Create records a new moderation action.
	Create(ctx context.Context, action *domain.Action) error
	// FindByID retrieves one action by identifier.
	FindByID(ctx context.Context, id int64) (*domain.Action, error)
	// List returns actions matching the filter.
	List(ctx context.Context, filter domain.ListFilter) ([]domain.Action, error)
	// Deactivate marks one action as inactive.
	Deactivate(ctx context.Context, id int64, deactivatedBy int) error
}

// TicketService defines ticket behavior for HTTP routes.
type TicketService interface {
	// Submit creates a new support ticket.
	Submit(ctx context.Context, ticket *domain.Ticket) error
	// FindByID retrieves one ticket.
	FindByID(ctx context.Context, id int64) (*domain.Ticket, error)
	// List returns tickets filtered by status.
	List(ctx context.Context, status domain.TicketStatus, limit int) ([]domain.Ticket, error)
	// Close resolves a ticket.
	Close(ctx context.Context, id int64, status domain.TicketStatus) error
}

// WordFilterService defines word filter behavior for HTTP routes.
type WordFilterService interface {
	// Create stores a new word filter rule.
	Create(ctx context.Context, entry *domain.WordFilterEntry) error
	// FindByID retrieves one word filter entry.
	FindByID(ctx context.Context, id int64) (*domain.WordFilterEntry, error)
	// ListActive returns active filters.
	ListActive(ctx context.Context, scope string, roomID int) ([]domain.WordFilterEntry, error)
	// Delete removes a word filter entry.
	Delete(ctx context.Context, id int64) error
}

// PresetService defines preset behavior for HTTP routes.
type PresetService interface {
	// Create stores a new preset.
	Create(ctx context.Context, preset *domain.Preset) error
	// ListActive returns all active presets.
	ListActive(ctx context.Context) ([]domain.Preset, error)
	// Delete removes a preset.
	Delete(ctx context.Context, id int64) error
}

// VisitService defines visit tracking behavior for HTTP routes.
type VisitService interface {
	// ListByUser returns recent visits for a user.
	ListByUser(ctx context.Context, userID int, limit int) ([]domain.VisitRecord, error)
	// ListByRoom returns recent visits for a room.
	ListByRoom(ctx context.Context, roomID int, limit int) ([]domain.VisitRecord, error)
}
