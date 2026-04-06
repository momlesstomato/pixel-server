package application

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// SoftDelete marks one room as deleted via the optional room repository.
func (s *Service) SoftDelete(ctx context.Context, roomID int) error {
	if s.rooms == nil {
		return domain.ErrRoomNotFound
	}
	return s.rooms.SoftDelete(ctx, roomID)
}

// CreateBan persists a new room ban entry.
func (s *Service) CreateBan(ctx context.Context, ban domain.RoomBan) (domain.RoomBan, error) {
	return s.bans.CreateBan(ctx, ban)
}

// ListBans returns all active bans for one room.
func (s *Service) ListBans(ctx context.Context, roomID int) ([]domain.RoomBan, error) {
	return s.bans.ListBansByRoom(ctx, roomID)
}

// FindBan resolves an active ban for one user in one room.
func (s *Service) FindBan(ctx context.Context, roomID, userID int) (*domain.RoomBan, error) {
	return s.bans.FindActiveBan(ctx, roomID, userID)
}

// RemoveBan deletes one room ban by identifier.
func (s *Service) RemoveBan(ctx context.Context, banID int) error {
	return s.bans.DeleteBan(ctx, banID)
}

// GrantRights grants room rights to one user.
func (s *Service) GrantRights(ctx context.Context, roomID, ownerID, targetID int) error {
	room, err := s.FindRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.OwnerID != ownerID {
		return domain.ErrAccessDenied
	}
	return s.rights.GrantRights(ctx, roomID, targetID)
}

// RevokeRights revokes room rights from one user.
func (s *Service) RevokeRights(ctx context.Context, roomID, ownerID, targetID int) error {
	room, err := s.FindRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.OwnerID != ownerID {
		return domain.ErrAccessDenied
	}
	return s.rights.RevokeRights(ctx, roomID, targetID)
}

// RevokeAllRights removes all rights holders from one room.
func (s *Service) RevokeAllRights(ctx context.Context, roomID, ownerID int) error {
	room, err := s.FindRoom(ctx, roomID)
	if err != nil {
		return err
	}
	if room.OwnerID != ownerID {
		return domain.ErrAccessDenied
	}
	return s.rights.RevokeAllRights(ctx, roomID)
}

// ListRights returns all rights holder user IDs for one room.
func (s *Service) ListRights(ctx context.Context, roomID, ownerID int) ([]int, error) {
	room, err := s.FindRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.OwnerID != ownerID {
		return nil, domain.ErrAccessDenied
	}
	return s.rights.ListRightsByRoom(ctx, roomID)
}
