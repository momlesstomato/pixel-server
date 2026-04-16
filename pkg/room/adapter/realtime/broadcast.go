package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	"go.uber.org/zap"
)

// publishToRoomEntities publishes one pre-encoded frame to every player entity in a room.
func (rt *Runtime) publishToRoomEntities(roomID int, frame []byte) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	ctx := context.Background()
	for _, entity := range inst.Entities() {
		if entity.Type != domain.EntityPlayer || entity.UserID == 0 {
			continue
		}
		ch := sessionnotification.UserChannel(entity.UserID)
		if err := rt.broadcaster.Publish(ctx, ch, frame); err != nil {
			rt.logger.Warn("publish to room entity failed", zap.Int("user_id", entity.UserID), zap.Error(err))
		}
	}
}

// Broadcast implements engine.EntityBroadcaster and sends dirty entity updates to all room players.
func (rt *Runtime) Broadcast(roomID int, updates []domain.RoomEntity, _ []byte) {
	body, err := packet.UserUpdateComposer{Entities: updates}.Encode()
	if err != nil {
		rt.logger.Warn("encode entity update failed", zap.Error(err))
		return
	}
	rt.publishToRoomEntities(roomID, codec.EncodeFrame(packet.UserUpdateComposerID, body))
}

// broadcastToRoom sends one packet to all player entities currently in a room.
func (rt *Runtime) broadcastToRoom(roomID int, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	body, err := pkt.Encode()
	if err != nil {
		rt.logger.Warn("encode broadcast packet failed", zap.Error(err))
		return
	}
	rt.publishToRoomEntities(roomID, codec.EncodeFrame(pkt.PacketID(), body))
}

// OnSleep broadcasts a sleep state change for one entity to all room players.
func (rt *Runtime) OnSleep(roomID int, virtualID int, sleeping bool) {
	rt.broadcastToRoom(roomID, packet.SleepComposer{VirtualID: int32(virtualID), IsAsleep: sleeping})
}

// BroadcastRawToRoom sends an already-encoded payload to all player entities in a room.
func (rt *Runtime) BroadcastRawToRoom(roomID int, packetID uint16, body []byte) {
	rt.publishToRoomEntities(roomID, codec.EncodeFrame(packetID, body))
}

// EjectSittingEntitiesInRoom clears the auto-sit state for entities at a tile, walks them
// toward the door, and broadcasts the update to all room players.
func (rt *Runtime) EjectSittingEntitiesInRoom(roomID, x, y int) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	updated := inst.EjectSittingEntitiesAt(x, y)
	if len(updated) == 0 {
		return
	}
	rt.Broadcast(roomID, updated, nil)
}

// RotateSittingEntitiesInRoom rotates auto-sitting entities at a tile to match the new furniture direction and broadcasts the update.
func (rt *Runtime) RotateSittingEntitiesInRoom(roomID, x, y, dir int) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	updated := inst.RotateSittingEntitiesAt(x, y, dir)
	if len(updated) == 0 {
		return
	}
	rt.Broadcast(roomID, updated, nil)
}

// ConnRoomID returns the room identifier for a given connection, if present.
func (rt *Runtime) ConnRoomID(connID string) (int, bool) {
	return rt.roomIDByConn(connID)
}

// ConnTile returns the active room tile for a given connection, if present.
func (rt *Runtime) ConnTile(connID string) (int, int, int, bool) {
	userID, ok := rt.userID(connID)
	if !ok {
		return 0, 0, 0, false
	}
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil || entity == nil {
		return 0, 0, 0, false
	}
	return inst.RoomID, entity.Position.X, entity.Position.Y, true
}

// OnKick broadcasts entity removal for one auto-kicked entity and removes connection tracking.
func (rt *Runtime) OnKick(roomID int, entity domain.RoomEntity) {
	body, err := packet.UserRemoveComposer{VirtualID: int32(entity.VirtualID)}.Encode()
	if err != nil {
		rt.logger.Warn("encode kick packet failed", zap.Error(err))
		return
	}
	rt.publishToRoomEntities(roomID, codec.EncodeFrame(packet.UserRemoveComposerID, body))
	rt.clearRoomForConn(entity.ConnID)
}

// OnDoorExit broadcasts entity removal when an entity walks out through the door tile,
// then publishes the hotel view redirect to the exiting user's broadcast channel.
func (rt *Runtime) OnDoorExit(roomID int, entity domain.RoomEntity) {
	body, err := packet.UserRemoveComposer{VirtualID: int32(entity.VirtualID)}.Encode()
	if err != nil {
		rt.logger.Warn("encode door exit packet failed", zap.Error(err))
		return
	}
	rt.publishToRoomEntities(roomID, codec.EncodeFrame(packet.UserRemoveComposerID, body))
	if entity.UserID > 0 {
		rt.publishPacketToUser(context.Background(), entity.UserID, sessionnavigation.DesktopViewResponsePacket{})
	}
	rt.clearRoomForConn(entity.ConnID)
}
