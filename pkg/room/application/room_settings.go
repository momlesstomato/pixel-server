package application

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
)

// ModerateSettings persists moderator-applied room settings without owner validation.
func (s *Service) ModerateSettings(ctx context.Context, updated domain.Room) error {
	if s.rooms == nil {
		return domain.ErrRoomNotFound
	}
	if updated.ID <= 0 {
		return domain.ErrRoomNotFound
	}
	if _, err := s.rooms.FindByID(ctx, updated.ID); err != nil {
		return err
	}
	return s.rooms.SaveSettings(ctx, updated)
}
