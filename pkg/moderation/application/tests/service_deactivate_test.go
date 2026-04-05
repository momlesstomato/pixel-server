package tests

import (
	"context"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/moderation/application"
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

func createService(t *testing.T) (*application.Service, *mockRepo) {
	t.Helper()
	repo := newMockRepo()
	svc, err := application.NewService(repo)
	require.NoError(t, err)
	return svc, repo
}

func seedAction(t *testing.T, svc *application.Service, scope domain.ActionScope, atype domain.ActionType, uid int) *domain.Action {
	t.Helper()
	a := &domain.Action{TargetUserID: uid, Scope: scope, ActionType: atype}
	require.NoError(t, svc.Create(context.Background(), a))
	return a
}

// TestDeactivateSuccess verifies deactivation of an active action.
func TestDeactivateSuccess(t *testing.T) {
	svc, _ := createService(t)
	a := seedAction(t, svc, domain.ScopeHotel, domain.TypeBan, 1)
	err := svc.Deactivate(context.Background(), a.ID, 99)
	assert.NoError(t, err)
}

// TestDeactivateAlreadyInactive verifies double deactivation error.
func TestDeactivateAlreadyInactive(t *testing.T) {
	svc, _ := createService(t)
	a := seedAction(t, svc, domain.ScopeHotel, domain.TypeBan, 1)
	require.NoError(t, svc.Deactivate(context.Background(), a.ID, 99))
	err := svc.Deactivate(context.Background(), a.ID, 99)
	assert.ErrorIs(t, err, domain.ErrAlreadyInactive)
}

// TestDeleteRoomAction verifies room action can be deleted.
func TestDeleteRoomAction(t *testing.T) {
	svc, _ := createService(t)
	a := seedAction(t, svc, domain.ScopeRoom, domain.TypeBan, 1)
	err := svc.Delete(context.Background(), a.ID)
	assert.NoError(t, err)
}

// TestDeleteHotelActionFails verifies hotel actions are not deletable.
func TestDeleteHotelActionFails(t *testing.T) {
	svc, _ := createService(t)
	a := seedAction(t, svc, domain.ScopeHotel, domain.TypeBan, 1)
	err := svc.Delete(context.Background(), a.ID)
	assert.ErrorIs(t, err, domain.ErrCannotDeleteHotelAction)
}

// TestListClampLimit verifies negative limit is clamped to 50.
func TestListClampLimit(t *testing.T) {
	svc, _ := createService(t)
	seedAction(t, svc, domain.ScopeHotel, domain.TypeBan, 1)
	result, err := svc.List(context.Background(), domain.ListFilter{Limit: -1})
	require.NoError(t, err)
	assert.Len(t, result, 1)
}
