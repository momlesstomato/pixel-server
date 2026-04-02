package application

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	sdknavigator "github.com/momlesstomato/pixel-sdk/events/navigator"
)

// ListRooms resolves paginated rooms with optional filter.
func (service *Service) ListRooms(ctx context.Context, filter domain.RoomFilter) ([]domain.Room, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return service.repository.ListRooms(ctx, filter)
}

// FindRoomByID resolves one room by identifier.
func (service *Service) FindRoomByID(ctx context.Context, id int) (domain.Room, error) {
	if id <= 0 {
		return domain.Room{}, fmt.Errorf("room id must be positive")
	}
	return service.repository.FindRoomByID(ctx, id)
}

// CreateRoom persists one validated room row.
func (service *Service) CreateRoom(ctx context.Context, room domain.Room) (domain.Room, error) {
	if room.Name == "" {
		return domain.Room{}, fmt.Errorf("room name is required")
	}
	if room.OwnerID <= 0 {
		return domain.Room{}, fmt.Errorf("owner id must be positive")
	}
	if service.fire != nil {
		ev := &sdknavigator.RoomCreating{OwnerID: room.OwnerID, Name: room.Name}
		service.fire(ev)
		if ev.Cancelled() {
			return domain.Room{}, fmt.Errorf("room creation cancelled by plugin")
		}
	}
	created, err := service.repository.CreateRoom(ctx, room)
	if err != nil {
		return domain.Room{}, err
	}
	if service.fire != nil {
		service.fire(&sdknavigator.RoomCreated{RoomID: created.ID, OwnerID: created.OwnerID, Name: created.Name})
	}
	return created, nil
}

// UpdateRoom applies partial room update.
func (service *Service) UpdateRoom(ctx context.Context, id int, patch domain.RoomPatch) (domain.Room, error) {
	if id <= 0 {
		return domain.Room{}, fmt.Errorf("room id must be positive")
	}
	return service.repository.UpdateRoom(ctx, id, patch)
}

// DeleteRoom removes one room by identifier.
func (service *Service) DeleteRoom(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("room id must be positive")
	}
	if service.fire != nil {
		ev := &sdknavigator.RoomDeleting{RoomID: id}
		service.fire(ev)
		if ev.Cancelled() {
			return fmt.Errorf("room deletion cancelled by plugin")
		}
	}
	if err := service.repository.DeleteRoom(ctx, id); err != nil {
		return err
	}
	if service.fire != nil {
		service.fire(&sdknavigator.RoomDeleted{RoomID: id})
	}
	return nil
}
