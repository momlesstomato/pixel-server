package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/room/packet"
)

// handleDance processes entity dance state change request.
func (rt *Runtime) handleDance(connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.DancePacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	if err := rt.entitySvc.Dance(context.Background(), inst, entity, int(pkt.DanceID)); err != nil {
		return nil
	}
	rt.broadcastToRoom(inst.RoomID, packet.DanceComposer{VirtualID: int32(entity.VirtualID), DanceID: pkt.DanceID})
	return nil
}

// handleAction processes entity action expression request.
func (rt *Runtime) handleAction(connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.ActionPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	return rt.entitySvc.Action(context.Background(), inst, entity, int(pkt.ActionID))
}

// handleSign processes entity sign display request.
func (rt *Runtime) handleSign(connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.SignPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	return rt.entitySvc.Sign(context.Background(), inst, entity, int(pkt.SignID))
}

// handleStartTyping sets entity typing indicator and broadcasts to room.
func (rt *Runtime) handleStartTyping(connID string, userID int) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	if err := rt.entitySvc.StartTyping(context.Background(), inst, entity); err != nil {
		return nil
	}
	rt.broadcastToRoom(inst.RoomID, packet.UserTypingComposer{VirtualID: int32(entity.VirtualID), IsTyping: true})
	return nil
}

// handleStopTyping clears entity typing indicator and broadcasts to room.
func (rt *Runtime) handleStopTyping(connID string, userID int) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	if err := rt.entitySvc.StopTyping(context.Background(), inst, entity); err != nil {
		return nil
	}
	rt.broadcastToRoom(inst.RoomID, packet.UserTypingComposer{VirtualID: int32(entity.VirtualID), IsTyping: false})
	return nil
}

// handleLookTo processes entity head rotation request.
func (rt *Runtime) handleLookTo(connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	var pkt packet.LookToPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if rt.entitySvc == nil {
		return nil
	}
	return rt.entitySvc.LookTo(context.Background(), inst, entity, int(pkt.X), int(pkt.Y))
}
