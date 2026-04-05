package store

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/model"
	"gorm.io/gorm"
)

// TicketStore implements domain.TicketRepository using GORM.
type TicketStore struct {
	// db stores the GORM database connection.
	db *gorm.DB
}

// NewTicketStore creates a ticket store backed by the given database.
func NewTicketStore(db *gorm.DB) (*TicketStore, error) {
	if db == nil {
		return nil, domain.ErrMissingTarget
	}
	return &TicketStore{db: db}, nil
}

// Create persists a new support ticket.
func (s *TicketStore) Create(ctx context.Context, ticket *domain.Ticket) error {
	m := ticketToModel(ticket)
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return err
	}
	ticket.ID = m.ID
	ticket.CreatedAt = m.CreatedAt
	return nil
}

// FindByID retrieves one ticket by identifier.
func (s *TicketStore) FindByID(ctx context.Context, id int64) (*domain.Ticket, error) {
	var m model.ModerationTicket
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, domain.ErrTicketNotFound
	}
	t := ticketToDomain(&m)
	return &t, nil
}

// List returns tickets filtered by status with a limit.
func (s *TicketStore) List(ctx context.Context, status domain.TicketStatus, limit int) ([]domain.Ticket, error) {
	q := s.db.WithContext(ctx).Model(&model.ModerationTicket{})
	if status != "" {
		q = q.Where("status = ?", string(status))
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	q = q.Order("created_at DESC")
	var rows []model.ModerationTicket
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Ticket, len(rows))
	for i := range rows {
		out[i] = ticketToDomain(&rows[i])
	}
	return out, nil
}

// UpdateStatus changes ticket status and optional assignee.
func (s *TicketStore) UpdateStatus(ctx context.Context, id int64, status domain.TicketStatus, assignedTo int) error {
	updates := map[string]interface{}{"status": string(status)}
	if assignedTo > 0 {
		updates["assigned_to"] = assignedTo
	}
	if status == domain.TicketClosed || status == domain.TicketInvalid || status == domain.TicketAbusive {
		now := time.Now()
		updates["closed_at"] = now
	}
	return s.db.WithContext(ctx).Model(&model.ModerationTicket{}).Where("id = ?", id).Updates(updates).Error
}

// Delete hard-deletes one ticket row.
func (s *TicketStore) Delete(ctx context.Context, id int64) error {
	return s.db.WithContext(ctx).Delete(&model.ModerationTicket{}, id).Error
}

func ticketToModel(t *domain.Ticket) model.ModerationTicket {
	return model.ModerationTicket{
		ReporterID: t.ReporterID, ReportedID: t.ReportedID, RoomID: t.RoomID,
		Category: t.Category, Message: t.Message, Status: string(t.Status), AssignedTo: t.AssignedTo,
	}
}

func ticketToDomain(m *model.ModerationTicket) domain.Ticket {
	return domain.Ticket{
		ID: m.ID, ReporterID: m.ReporterID, ReportedID: m.ReportedID, RoomID: m.RoomID,
		Category: m.Category, Message: m.Message, Status: domain.TicketStatus(m.Status),
		AssignedTo: m.AssignedTo, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, ClosedAt: m.ClosedAt,
	}
}
