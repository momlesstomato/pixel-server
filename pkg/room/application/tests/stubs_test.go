package tests

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// modelRepoStub provides deterministic model lookup for tests.
type modelRepoStub struct{ models map[string]domain.RoomModel }

// FindModelBySlug resolves one room model by slug identifier.
func (m *modelRepoStub) FindModelBySlug(_ context.Context, slug string) (domain.RoomModel, error) {
	model, ok := m.models[slug]
	if !ok {
		return domain.RoomModel{}, domain.ErrRoomModelNotFound
	}
	return model, nil
}

// ListModels returns all room model templates.
func (m *modelRepoStub) ListModels(_ context.Context) ([]domain.RoomModel, error) {
	out := make([]domain.RoomModel, 0, len(m.models))
	for _, v := range m.models {
		out = append(out, v)
	}
	return out, nil
}

// banRepoStub provides deterministic ban lookup for tests.
type banRepoStub struct{ banned map[[2]int]bool }

// FindActiveBan returns a ban if the pair is marked banned.
func (m *banRepoStub) FindActiveBan(_ context.Context, roomID, userID int) (*domain.RoomBan, error) {
	if m.banned != nil && m.banned[[2]int{roomID, userID}] {
		return &domain.RoomBan{RoomID: roomID, UserID: userID}, nil
	}
	return nil, nil
}

// CreateBan persists one room ban entry.
func (m *banRepoStub) CreateBan(_ context.Context, b domain.RoomBan) (domain.RoomBan, error) {
	return b, nil
}

// DeleteBan removes one room ban by identifier.
func (m *banRepoStub) DeleteBan(_ context.Context, _ int) error { return nil }

// ListBansByRoom returns all bans for one room.
func (m *banRepoStub) ListBansByRoom(_ context.Context, _ int) ([]domain.RoomBan, error) {
	return nil, nil
}

// rightsRepoStub provides deterministic rights lookup for tests.
type rightsRepoStub struct{ rights map[[2]int]bool }

// HasRights checks if a user has rights in a room.
func (m *rightsRepoStub) HasRights(_ context.Context, roomID, userID int) (bool, error) {
	if m.rights == nil {
		return false, nil
	}
	return m.rights[[2]int{roomID, userID}], nil
}

// GrantRights adds rights for one user in one room.
func (m *rightsRepoStub) GrantRights(_ context.Context, _, _ int) error { return nil }

// RevokeRights removes rights for one user in one room.
func (m *rightsRepoStub) RevokeRights(_ context.Context, _, _ int) error { return nil }

// ListRightsByRoom returns all rights holders for one room.
func (m *rightsRepoStub) ListRightsByRoom(_ context.Context, _ int) ([]int, error) { return nil, nil }

// RevokeAllRights removes all rights for one room.
func (m *rightsRepoStub) RevokeAllRights(_ context.Context, _ int) error { return nil }

// roomRepoStub provides deterministic room data persistence for tests.
type roomRepoStub struct{ rooms map[int]domain.Room }

// FindByID resolves one room record by identifier.
func (s *roomRepoStub) FindByID(_ context.Context, id int) (domain.Room, error) {
	r, ok := s.rooms[id]
	if !ok {
		return domain.Room{}, domain.ErrRoomNotFound
	}
	return r, nil
}

// SaveSettings persists updated room settings.
func (s *roomRepoStub) SaveSettings(_ context.Context, _ domain.Room) error { return nil }
