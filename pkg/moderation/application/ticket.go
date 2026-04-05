package application

import (
	"context"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
)

// TicketService implements call-for-help ticket business logic.
type TicketService struct {
	// repo stores the ticket persistence layer.
	repo domain.TicketRepository
	// fire dispatches SDK events when configured.
	fire func(sdk.Event)
}

// NewTicketService creates a ticket service.
func NewTicketService(repo domain.TicketRepository) (*TicketService, error) {
	if repo == nil {
		return nil, domain.ErrMissingTarget
	}
	return &TicketService{repo: repo}, nil
}

// SetEventFirer configures the SDK event dispatcher.
func (s *TicketService) SetEventFirer(fn func(sdk.Event)) {
	s.fire = fn
}

// Submit creates a new support ticket.
func (s *TicketService) Submit(ctx context.Context, ticket *domain.Ticket) error {
	if ticket.ReporterID <= 0 {
		return domain.ErrMissingTarget
	}
	ticket.Status = domain.TicketOpen
	return s.repo.Create(ctx, ticket)
}

// FindByID retrieves one ticket.
func (s *TicketService) FindByID(ctx context.Context, id int64) (*domain.Ticket, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns tickets filtered by status.
func (s *TicketService) List(ctx context.Context, status domain.TicketStatus, limit int) ([]domain.Ticket, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repo.List(ctx, status, limit)
}

// Assign moves a ticket to in-progress with a moderator.
func (s *TicketService) Assign(ctx context.Context, id int64, moderatorID int) error {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if ticket.Status == domain.TicketClosed || ticket.Status == domain.TicketInvalid {
		return domain.ErrTicketAlreadyClosed
	}
	return s.repo.UpdateStatus(ctx, id, domain.TicketInProgress, moderatorID)
}

// Close resolves a ticket with the given resolution status.
func (s *TicketService) Close(ctx context.Context, id int64, status domain.TicketStatus) error {
	ticket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if ticket.Status == domain.TicketClosed {
		return domain.ErrTicketAlreadyClosed
	}
	if status != domain.TicketClosed && status != domain.TicketInvalid && status != domain.TicketAbusive {
		status = domain.TicketClosed
	}
	return s.repo.UpdateStatus(ctx, id, status, 0)
}

// Delete removes a ticket.
func (s *TicketService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
