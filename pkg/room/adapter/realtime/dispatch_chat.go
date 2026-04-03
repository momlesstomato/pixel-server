package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	"go.uber.org/zap"
)

// handleChat processes proximity talk message request.
func (rt *Runtime) handleChat(ctx context.Context, connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.ChatPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.chatSvc == nil {
		return nil
	}
	recipients, err := rt.chatSvc.Talk(ctx, inst, entity, inst.RoomID, pkt.Message, int(pkt.BubbleStyle))
	if err != nil {
		rt.logger.Debug("talk rejected", zap.String("conn_id", connID), zap.Error(err))
		return nil
	}
	composer := packet.ChatComposer{VirtualID: int32(entity.VirtualID), Message: pkt.Message, BubbleStyle: pkt.BubbleStyle}
	rt.sendToRecipients(recipients, composer)
	return nil
}

// handleShout processes room-wide shout message request.
func (rt *Runtime) handleShout(ctx context.Context, connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.ShoutPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.chatSvc == nil {
		return nil
	}
	recipients, err := rt.chatSvc.Shout(ctx, inst, entity, inst.RoomID, pkt.Message, int(pkt.BubbleStyle))
	if err != nil {
		rt.logger.Debug("shout rejected", zap.String("conn_id", connID), zap.Error(err))
		return nil
	}
	composer := packet.ShoutComposer{VirtualID: int32(entity.VirtualID), Message: pkt.Message, BubbleStyle: pkt.BubbleStyle}
	rt.sendToRecipients(recipients, composer)
	return nil
}

// handleWhisper processes targeted private message request.
func (rt *Runtime) handleWhisper(ctx context.Context, connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.WhisperPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.chatSvc == nil {
		return nil
	}
	entities := inst.Entities()
	var target *domain.RoomEntity
	for i := range entities {
		if entities[i].Username == pkt.TargetUsername {
			target = &entities[i]
			break
		}
	}
	if target == nil {
		return nil
	}
	recipients, err := rt.chatSvc.Whisper(ctx, entity, inst.RoomID, target, pkt.Message, int(pkt.BubbleStyle))
	if err != nil {
		rt.logger.Debug("whisper rejected", zap.String("conn_id", connID), zap.Error(err))
		return nil
	}
	composer := packet.WhisperComposer{VirtualID: int32(entity.VirtualID), SenderName: entity.Username, Message: pkt.Message, BubbleStyle: pkt.BubbleStyle}
	rt.sendToRecipients(recipients, composer)
	return nil
}

// sendToRecipients transmits one encoded packet to a list of player entities.
func (rt *Runtime) sendToRecipients(recipients []domain.RoomEntity, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	body, err := pkt.Encode()
	if err != nil {
		rt.logger.Warn("encode recipient packet failed", zap.Error(err))
		return
	}
	for i := range recipients {
		if recipients[i].Type != domain.EntityPlayer || recipients[i].ConnID == "" {
			continue
		}
		if err := rt.transport.Send(recipients[i].ConnID, pkt.PacketID(), body); err != nil {
			rt.logger.Warn("send to recipient failed", zap.String("conn_id", recipients[i].ConnID), zap.Error(err))
		}
	}
}
