package realtime

import (
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"go.uber.org/zap"
)

// Broadcast implements engine.EntityBroadcaster and sends dirty entity updates to all room players.
func (rt *Runtime) Broadcast(roomID int, updates []domain.RoomEntity, _ []byte) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	body, err := packet.UserUpdateComposer{Entities: updates}.Encode()
	if err != nil {
		rt.logger.Warn("encode entity update failed", zap.Error(err))
		return
	}
	for _, entity := range inst.Entities() {
		if entity.Type != domain.EntityPlayer || entity.ConnID == "" {
			continue
		}
		if err := rt.transport.Send(entity.ConnID, packet.UserUpdateComposerID, body); err != nil {
			rt.logger.Warn("send entity update failed", zap.String("conn_id", entity.ConnID), zap.Error(err))
		}
	}
}

// broadcastToRoom sends one packet to all player entities currently in a room.
func (rt *Runtime) broadcastToRoom(roomID int, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	body, err := pkt.Encode()
	if err != nil {
		rt.logger.Warn("encode broadcast packet failed", zap.Error(err))
		return
	}
	for _, entity := range inst.Entities() {
		if entity.Type != domain.EntityPlayer || entity.ConnID == "" {
			continue
		}
		if err := rt.transport.Send(entity.ConnID, pkt.PacketID(), body); err != nil {
			rt.logger.Warn("send broadcast packet failed", zap.String("conn_id", entity.ConnID), zap.Error(err))
		}
	}
}

// OnSleep broadcasts a sleep state change for one entity to all room players.
func (rt *Runtime) OnSleep(roomID int, virtualID int, sleeping bool) {
	rt.broadcastToRoom(roomID, packet.SleepComposer{VirtualID: int32(virtualID), IsAsleep: sleeping})
}

// BroadcastRawToRoom sends an already-encoded payload to all player entities in a room.
func (rt *Runtime) BroadcastRawToRoom(roomID int, packetID uint16, body []byte) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	for _, entity := range inst.Entities() {
		if entity.Type != domain.EntityPlayer || entity.ConnID == "" {
			continue
		}
		if err := rt.transport.Send(entity.ConnID, packetID, body); err != nil {
			rt.logger.Warn("broadcast raw to room failed", zap.String("conn_id", entity.ConnID), zap.Error(err))
		}
	}
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
	id, ok := rt.connRooms[connID]
	return id, ok
}

// OnKick broadcasts entity removal for one auto-kicked entity and removes connection tracking.
func (rt *Runtime) OnKick(roomID int, entity domain.RoomEntity) {
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return
	}
	body, err := packet.UserRemoveComposer{VirtualID: int32(entity.VirtualID)}.Encode()
	if err != nil {
		rt.logger.Warn("encode kick packet failed", zap.Error(err))
		return
	}
	for _, e := range inst.Entities() {
		if e.Type != domain.EntityPlayer || e.ConnID == "" {
			continue
		}
		if err := rt.transport.Send(e.ConnID, packet.UserRemoveComposerID, body); err != nil {
			rt.logger.Warn("send kick packet failed", zap.String("conn_id", e.ConnID), zap.Error(err))
		}
	}
	delete(rt.connRooms, entity.ConnID)
}
