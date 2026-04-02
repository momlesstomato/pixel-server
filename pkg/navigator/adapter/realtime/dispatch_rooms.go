package realtime

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/navigator/application"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	"github.com/momlesstomato/pixel-server/pkg/navigator/packet"
	"go.uber.org/zap"
)

// handleGetGuestRoom responds with detailed room data.
func (runtime *Runtime) handleGetGuestRoom(ctx context.Context, connID string, body []byte) error {
	id := parseRoomID(body)
	if id <= 0 {
		return nil
	}
	room, err := runtime.service.FindRoomByID(ctx, id)
	if err != nil {
		if !errors.Is(err, domain.ErrRoomNotFound) {
			runtime.logger.Error("get guest room failed", zap.Int("room_id", id), zap.Error(err))
		}
		return nil
	}
	return runtime.sendPacket(connID, packet.GuestRoomDataPacket{Room: room, Forward: false})
}

// handleCreateRoom processes room creation request.
func (runtime *Runtime) handleCreateRoom(ctx context.Context, connID string, userID int, body []byte) error {
	name, desc, state, catID, maxUsers, tags := parseCreateRoom(body)
	room, err := runtime.service.CreateRoom(ctx, domain.Room{
		OwnerID: userID, Name: name, Description: desc, State: state,
		CategoryID: catID, MaxUsers: maxUsers, Tags: tags,
	})
	if err != nil {
		runtime.logger.Error("create room failed", zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	return runtime.sendPacket(connID, packet.RoomCreatedPacket{RoomID: int32(room.ID), Name: room.Name})
}

// handleAddFavourite adds one room to user favourites.
func (runtime *Runtime) handleAddFavourite(ctx context.Context, connID string, userID int, body []byte) error {
	roomID := parseRoomID(body)
	if roomID <= 0 {
		return nil
	}
	if err := runtime.service.AddFavourite(ctx, userID, roomID); err != nil {
		if !errors.Is(err, domain.ErrFavouriteLimitReached) && !errors.Is(err, domain.ErrFavouriteAlreadyExists) {
			runtime.logger.Error("add favourite failed", zap.Int("user_id", userID), zap.Error(err))
		}
		return nil
	}
	if err := runtime.sendPacket(connID, packet.FavouriteChangedPacket{RoomID: int32(roomID), Added: true}); err != nil {
		return err
	}
	return runtime.sendFavourites(ctx, connID, userID)
}

// handleRemoveFavourite removes one room from user favourites.
func (runtime *Runtime) handleRemoveFavourite(ctx context.Context, connID string, userID int, body []byte) error {
	roomID := parseRoomID(body)
	if roomID <= 0 {
		return nil
	}
	if err := runtime.service.RemoveFavourite(ctx, userID, roomID); err != nil {
		runtime.logger.Error("remove favourite failed", zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	if err := runtime.sendPacket(connID, packet.FavouriteChangedPacket{RoomID: int32(roomID), Added: false}); err != nil {
		return err
	}
	return runtime.sendFavourites(ctx, connID, userID)
}

// sendFavourites sends updated favourites list to one connection.
func (runtime *Runtime) sendFavourites(ctx context.Context, connID string, userID int) error {
	favs, err := runtime.service.ListFavourites(ctx, userID)
	if err != nil {
		return nil
	}
	ids := make([]int32, len(favs))
	for i, f := range favs {
		ids[i] = int32(f.RoomID)
	}
	return runtime.sendPacket(connID, packet.FavouritesListPacket{
		MaxFavourites: int32(application.MaxFavourites), RoomIDs: ids,
	})
}

// parseCreateRoom reads room creation fields from packet body.
func parseCreateRoom(body []byte) (string, string, string, int, int, []string) {
	r := codec.NewReader(body)
	name, _ := r.ReadString()
	desc, _ := r.ReadString()
	model, _ := r.ReadString()
	catID, _ := r.ReadInt32()
	maxUsers, _ := r.ReadInt32()
	tradeMode, _ := r.ReadInt32()
	tagCount, _ := r.ReadInt32()
	tags := make([]string, 0, tagCount)
	for i := int32(0); i < tagCount; i++ {
		tag, err := r.ReadString()
		if err != nil {
			break
		}
		tags = append(tags, tag)
	}
	_ = model
	_ = tradeMode
	return name, desc, "open", int(catID), int(maxUsers), tags
}
