package application

import (
	"context"
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	actions map[int64]*domain.Action
	nextID  int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{actions: make(map[int64]*domain.Action), nextID: 1}
}

func (m *mockRepo) Create(_ context.Context, a *domain.Action) error {
	a.ID = m.nextID
	m.nextID++
	cp := *a
	m.actions[a.ID] = &cp
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id int64) (*domain.Action, error) {
	a, ok := m.actions[id]
	if !ok {
		return nil, domain.ErrActionNotFound
	}
	return a, nil
}

func (m *mockRepo) List(_ context.Context, f domain.ListFilter) ([]domain.Action, error) {
	var out []domain.Action
	for _, a := range m.actions {
		if f.TargetUserID > 0 && a.TargetUserID != f.TargetUserID {
			continue
		}
		out = append(out, *a)
	}
	return out, nil
}

func (m *mockRepo) Deactivate(_ context.Context, id int64, by int) error {
	a, ok := m.actions[id]
	if !ok {
		return domain.ErrActionNotFound
	}
	a.Active = false
	a.DeactivatedBy = by
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id int64) error {
	delete(m.actions, id)
	return nil
}

func (m *mockRepo) HasActiveBan(_ context.Context, uid int, scope domain.ActionScope) (bool, error) {
	for _, a := range m.actions {
		if a.TargetUserID == uid && a.Scope == scope && a.ActionType == domain.TypeBan && a.Active {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) HasActiveMute(_ context.Context, uid int, scope domain.ActionScope) (bool, error) {
	for _, a := range m.actions {
		if a.TargetUserID == uid && a.Scope == scope && a.ActionType == domain.TypeMute && a.Active {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) HasActiveIPBan(_ context.Context, ip string) (bool, error) {
	for _, a := range m.actions {
		if a.IPAddress == ip && a.ActionType == domain.TypeBan && a.Active {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) HasActiveTradeLock(_ context.Context, uid int) (bool, error) {
	for _, a := range m.actions {
		if a.TargetUserID == uid && a.ActionType == domain.TypeTradeLock && a.Active {
			return true, nil
		}
	}
	return false, nil
}

// TestNewServiceNilRepo verifies nil repo is rejected.
func TestNewServiceNilRepo(t *testing.T) {
	_, err := NewService(nil)
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestNewServiceValid verifies service creation succeeds.
func TestNewServiceValid(t *testing.T) {
	svc, err := NewService(newMockRepo())
	require.NoError(t, err)
	assert.NotNil(t, svc)
}

// TestCreateMissingTarget verifies zero target user is rejected.
func TestCreateMissingTarget(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	err := svc.Create(context.Background(), &domain.Action{Scope: domain.ScopeHotel})
	assert.ErrorIs(t, err, domain.ErrMissingTarget)
}

// TestCreateInvalidScope verifies unknown scope is rejected.
func TestCreateInvalidScope(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	err := svc.Create(context.Background(), &domain.Action{TargetUserID: 1, Scope: "bad"})
	assert.ErrorIs(t, err, domain.ErrInvalidScope)
}

// TestCreateSuccess verifies valid action is persisted.
func TestCreateSuccess(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	action := &domain.Action{TargetUserID: 1, Scope: domain.ScopeHotel, ActionType: domain.TypeBan}
	err := svc.Create(context.Background(), action)
	require.NoError(t, err)
	assert.True(t, action.Active)
	assert.Equal(t, int64(1), action.ID)
}

// TestCreateWithDuration verifies ExpiresAt is computed.
func TestCreateWithDuration(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	action := &domain.Action{TargetUserID: 1, Scope: domain.ScopeHotel, ActionType: domain.TypeMute, DurationMinutes: 30}
	err := svc.Create(context.Background(), action)
	require.NoError(t, err)
	assert.NotNil(t, action.ExpiresAt)
}

// TestIsTradeLocked verifies trade lock lookup.
func TestIsTradeLocked(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	locked, err := svc.IsTradeLocked(context.Background(), 1)
	require.NoError(t, err)
	assert.False(t, locked)
}

// TestEscalateNoHistory verifies warn for clean user.
func TestEscalateNoHistory(t *testing.T) {
	svc, _ := NewService(newMockRepo())
	action, err := svc.Escalate(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, domain.TypeWarn, action.ActionType)
}

// TestEscalateWithHistory verifies escalation tiers.
func TestEscalateWithHistory(t *testing.T) {
	repo := newMockRepo()
	svc, _ := NewService(repo)
	ctx := context.Background()
	_ = svc.Create(ctx, &domain.Action{TargetUserID: 1, Scope: domain.ScopeHotel, ActionType: domain.TypeWarn})
	action, err := svc.Escalate(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, domain.TypeMute, action.ActionType)
}

// suppress unused import
var _ sdk.Event
