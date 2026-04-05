package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/moderation/application"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTicketRepo struct {
	tickets map[int64]*domain.Ticket
	nextID  int64
}

func newMockTicketRepo() *mockTicketRepo {
	return &mockTicketRepo{tickets: make(map[int64]*domain.Ticket), nextID: 1}
}

func (m *mockTicketRepo) Create(_ context.Context, t *domain.Ticket) error {
	t.ID = m.nextID
	m.nextID++
	cp := *t
	m.tickets[t.ID] = &cp
	return nil
}

func (m *mockTicketRepo) FindByID(_ context.Context, id int64) (*domain.Ticket, error) {
	t, ok := m.tickets[id]
	if !ok {
		return nil, domain.ErrTicketNotFound
	}
	return t, nil
}

func (m *mockTicketRepo) List(_ context.Context, status domain.TicketStatus, limit int) ([]domain.Ticket, error) {
	var out []domain.Ticket
	for _, t := range m.tickets {
		if status != "" && t.Status != status {
			continue
		}
		out = append(out, *t)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m *mockTicketRepo) UpdateStatus(_ context.Context, id int64, status domain.TicketStatus, _ int) error {
	t, ok := m.tickets[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	t.Status = status
	return nil
}

func (m *mockTicketRepo) Delete(_ context.Context, id int64) error {
	delete(m.tickets, id)
	return nil
}

// TestNewTicketServiceNilRepo verifies nil repo is rejected.
func TestNewTicketServiceNilRepo(t *testing.T) {
	_, err := application.NewTicketService(nil)
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestTicketSubmitMissingReporter verifies missing reporter is rejected.
func TestTicketSubmitMissingReporter(t *testing.T) {
	svc, _ := application.NewTicketService(newMockTicketRepo())
	err := svc.Submit(context.Background(), &domain.Ticket{})
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestTicketSubmitSuccess verifies successful ticket submission.
func TestTicketSubmitSuccess(t *testing.T) {
	svc, _ := application.NewTicketService(newMockTicketRepo())
	ticket := &domain.Ticket{ReporterID: 1, ReportedID: 2, Message: "test"}
	require.NoError(t, svc.Submit(context.Background(), ticket))
	assert.Equal(t, domain.TicketOpen, ticket.Status)
	assert.Greater(t, ticket.ID, int64(0))
}

// TestTicketListLimitClamp verifies list limit clamping.
func TestTicketListLimitClamp(t *testing.T) {
	svc, _ := application.NewTicketService(newMockTicketRepo())
	tickets, err := svc.List(context.Background(), "", 0)
	require.NoError(t, err)
	assert.Empty(t, tickets)
}

// TestTicketAssignClosedTicket verifies cannot assign a closed ticket.
func TestTicketAssignClosedTicket(t *testing.T) {
	repo := newMockTicketRepo()
	svc, _ := application.NewTicketService(repo)
	ticket := &domain.Ticket{ReporterID: 1, Message: "test"}
	_ = svc.Submit(context.Background(), ticket)
	_ = svc.Close(context.Background(), ticket.ID, domain.TicketClosed)
	err := svc.Assign(context.Background(), ticket.ID, 99)
	assert.ErrorIs(t, err, domain.ErrTicketAlreadyClosed)
}

// TestTicketCloseAlreadyClosed verifies double close is rejected.
func TestTicketCloseAlreadyClosed(t *testing.T) {
	repo := newMockTicketRepo()
	svc, _ := application.NewTicketService(repo)
	ticket := &domain.Ticket{ReporterID: 1, Message: "test"}
	_ = svc.Submit(context.Background(), ticket)
	_ = svc.Close(context.Background(), ticket.ID, domain.TicketClosed)
	err := svc.Close(context.Background(), ticket.ID, domain.TicketClosed)
	assert.ErrorIs(t, err, domain.ErrTicketAlreadyClosed)
}

// TestTicketCloseInvalidStatus verifies invalid status defaults to closed.
func TestTicketCloseInvalidStatus(t *testing.T) {
	repo := newMockTicketRepo()
	svc, _ := application.NewTicketService(repo)
	ticket := &domain.Ticket{ReporterID: 1, Message: "test"}
	_ = svc.Submit(context.Background(), ticket)
	require.NoError(t, svc.Close(context.Background(), ticket.ID, "unknown"))
	found, _ := svc.FindByID(context.Background(), ticket.ID)
	assert.Equal(t, domain.TicketClosed, found.Status)
}
