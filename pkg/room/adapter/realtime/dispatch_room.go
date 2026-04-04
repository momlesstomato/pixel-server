package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"go.uber.org/zap"
)

// triggerDoorbell sends a doorbell notification to the room owner and parks the visitor.
func (rt *Runtime) triggerDoorbell(_ context.Context, connID string, username string, room domain.Room) error {
	rt.pendingDoorbell[username] = doorbellEntry{connID: connID, roomID: room.ID}
	inst, ok := rt.service.Manager().Get(room.ID)
	if !ok {
		return rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	for _, entity := range inst.Entities() {
		if entity.UserID != room.OwnerID {
			continue
		}
		if err := rt.sendPacket(entity.ConnID, packet.DoorbellComposer{Username: username}); err != nil {
			rt.logger.Warn("doorbell notify failed", zap.String("conn_id", entity.ConnID), zap.Error(err))
		}
		return nil
	}
	delete(rt.pendingDoorbell, username)
	return rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 1})
}

// handleLetUserIn processes room owner doorbell approval or denial.
func (rt *Runtime) handleLetUserIn(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.LetUserInPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	entry, ok := rt.pendingDoorbell[pkt.Username]
	if !ok {
		return nil
	}
	delete(rt.pendingDoorbell, pkt.Username)
	if !pkt.Let {
		return rt.sendPacket(entry.connID, packet.FlatAccessibleComposer{Username: pkt.Username, Accessible: false})
	}
	roomID, ownerRoomID := rt.connRooms[connID]
	if !ownerRoomID || roomID != entry.roomID {
		return rt.sendPacket(entry.connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	room, err := rt.service.FindRoom(ctx, entry.roomID)
	if err != nil {
		return rt.sendPacket(entry.connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	inst, ok := rt.service.Manager().Get(entry.roomID)
	if !ok {
		return rt.sendPacket(entry.connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	if err := rt.sendPacket(entry.connID, packet.FlatAccessibleComposer{Username: pkt.Username, Accessible: true}); err != nil {
		return err
	}
	if err := rt.sendPacket(entry.connID, packet.OpenConnectionComposer{}); err != nil {
		return err
	}
	rt.connRooms[entry.connID] = entry.roomID
	visitorID, visitorFound := rt.userID(entry.connID)
	if !visitorFound {
		return nil
	}
	if err := rt.sendRoomData(entry.connID, visitorID, inst, room); err != nil {
		return err
	}
	return nil
}

// handleGetRoomSettings sends room settings to the requesting owner.
func (rt *Runtime) handleGetRoomSettings(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.GetRoomSettingsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, int(pkt.RoomID))
	if err != nil {
		return nil
	}
	if room.OwnerID != userID {
		return nil
	}
	return rt.sendPacket(connID, packet.RoomSettingsComposer{Room: room})
}

// handleSaveRoomSettings persists updated room settings from the owner.
func (rt *Runtime) handleSaveRoomSettings(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.SaveRoomSettingsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	updated := domain.Room{
		Name: pkt.Name, Description: pkt.Description,
		State:          packet.IntToAccessState(pkt.State),
		Password:       pkt.Password,
		MaxUsers:       int(pkt.MaxUsers),
		AllowPets:      pkt.AllowPets,
		AllowTrading:   pkt.AllowTrading,
		TradeMode:      int(pkt.TradeMode),
		WallThickness:  int(pkt.WallThickness),
		FloorThickness: int(pkt.FloorThickness),
		WallHeight:     int(pkt.WallHeight),
	}
	if err := rt.service.SaveSettings(ctx, int(pkt.RoomID), userID, updated); err != nil {
		rt.logger.Warn("save room settings failed", zap.Int("room_id", int(pkt.RoomID)), zap.Error(err))
		return nil
	}
	return rt.sendPacket(connID, packet.RoomSettingsSavedComposer{RoomID: pkt.RoomID})
}

// handleLeaveRoom removes the connection from the current room.
// Client-initiated departures (hotel view button, close flat) handle their own navigation.
func (rt *Runtime) handleLeaveRoom(connID string) error {
	rt.leaveCurrentRoom(connID)
	return nil
}
