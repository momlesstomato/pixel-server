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
