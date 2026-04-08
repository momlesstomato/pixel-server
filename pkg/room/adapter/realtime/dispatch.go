package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/engine"
	"github.com/momlesstomato/pixel-server/pkg/room/heightmap"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated room packet payload.
func (rt *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := rt.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.OpenFlatConnectionPacketID:
		return true, rt.handleOpenFlat(ctx, connID, userID, body)
	case packet.GetRoomEntryDataPacketID:
		return true, rt.handleGetEntryData(ctx, connID, userID)
	case packet.MoveAvatarPacketID:
		return true, rt.handleMoveAvatar(connID, userID, body)
	case packet.ChatPacketID:
		return true, rt.handleChat(ctx, connID, userID, body)
	case packet.ShoutPacketID:
		return true, rt.handleShout(ctx, connID, userID, body)
	case packet.WhisperPacketID:
		return true, rt.handleWhisper(ctx, connID, userID, body)
	case packet.DancePacketID:
		return true, rt.handleDance(connID, userID, body)
	case packet.ActionPacketID:
		return true, rt.handleAction(connID, userID, body)
	case packet.SignPacketID:
		return true, rt.handleSign(connID, userID, body)
	case packet.StartTypingPacketID:
		return true, rt.handleStartTyping(connID, userID)
	case packet.CancelTypingPacketID:
		return true, rt.handleStopTyping(connID, userID)
	case packet.LookToPacketID:
		return true, rt.handleLookTo(connID, userID, body)
	case packet.SitPacketID:
		return true, rt.handleSit(connID, userID)
	case packet.LetUserInPacketID:
		return true, rt.handleLetUserIn(ctx, connID, userID, body)
	case packet.KickUserPacketID:
		return true, rt.handleKickUser(ctx, connID, userID, body)
	case packet.BanUserPacketID:
		return true, rt.handleBanUser(ctx, connID, userID, body)
	case packet.GetRoomSettingsPacketID:
		return true, rt.handleGetRoomSettings(ctx, connID, userID, body)
	case packet.SaveRoomSettingsPacketID:
		return true, rt.handleSaveRoomSettings(ctx, connID, userID, body)
	case packet.GiveRoomScorePacketID:
		return true, rt.handleGiveRoomScore(ctx, connID, userID, body)
	case packet.DeleteRoomPacketID:
		return true, rt.handleDeleteRoom(ctx, connID, userID, body)
	case packet.GetBannedUsersPacketID:
		return true, rt.handleGetBannedUsers(ctx, connID, userID, body)
	case packet.UnbanUserPacketID:
		return rt.handleUnbanOrPass(ctx, connID, userID, body)
	case packet.AssignRightsPacketID:
		return true, rt.handleAssignRights(ctx, connID, userID, body)
	case packet.RemoveRightsPacketID:
		return true, rt.handleRemoveRights(ctx, connID, userID, body)
	case packet.RemoveMyRightsPacketID:
		return true, rt.handleRemoveMyRights(ctx, connID, userID)
	case packet.RemoveAllRightsPacketID:
		return true, rt.handleRemoveAllRights(ctx, connID, userID)
	case packet.GetRoomRightsPacketID:
		return true, rt.handleGetRoomRights(ctx, connID, userID)
	case packet.ToggleMuteToolPacketID:
		return true, rt.handleToggleMuteTool(connID, userID)
	case packet.RoomMuteUserPacketID:
		return true, rt.handleRoomMuteUser(ctx, connID, userID, body)
	case packet.DesktopViewPacketID, packet.CloseConnectionPacketID:
		return true, rt.handleLeaveRoom(connID)
	default:
		return false, nil
	}
}

// handleOpenFlat processes room entry request, checking access state.
func (rt *Runtime) handleOpenFlat(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.OpenFlatConnectionPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID := int(pkt.RoomID)
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil {
		rt.logger.Warn("room lookup failed", zap.Int("room_id", roomID), zap.Error(err))
		return rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	if room.ForwardRoomID > 0 {
		return rt.sendPacket(connID, packet.RoomForwardComposer{RoomID: int32(room.ForwardRoomID)})
	}
	canBypass := rt.canBypassRoomAccess(ctx, userID, room)
	if room.State == domain.AccessPassword && !canBypass {
		if retryAfter, active := rt.currentCooldown(userID, roomID); active {
			return rt.sendPasswordCooldownFeedback(connID, retryAfter)
		}
	}
	if accessErr := rt.service.CheckAccess(ctx, room, pkt.Password, userID); accessErr != nil {
		if accessErr == domain.ErrRoomBanned {
			return rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 4})
		}
		if accessErr == domain.ErrInvalidPassword {
			if canBypass {
				goto allowEntry
			}
			if retryAfter, cooldown := rt.recordPasswordFailure(userID, roomID); cooldown {
				return rt.sendPasswordCooldownFeedback(connID, retryAfter)
			}
			return rt.sendWrongPasswordFeedback(connID)
		}
		if accessErr == domain.ErrAccessDenied && canBypass {
			goto allowEntry
		}
		username := rt.resolveUsername(ctx, userID)
		return rt.triggerDoorbell(ctx, connID, userID, username, room)
	}
allowEntry:
	rt.clearPasswordFailures(userID, roomID)
	inst, err := rt.service.LoadRoom(ctx, room)
	if err != nil {
		rt.logger.Warn("room load failed", zap.Int("room_id", roomID), zap.Error(err))
		return rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 1})
	}
	if err := rt.sendPacket(connID, packet.OpenConnectionComposer{}); err != nil {
		return err
	}
	rt.leaveCurrentRoom(connID)
	rt.setRoomForConn(connID, roomID)
	if rt.visitRecorder != nil {
		_ = rt.visitRecorder.RecordVisit(ctx, userID, roomID)
	}
	return rt.sendRoomData(connID, userID, inst, room)
}

// handleGetEntryData sends room geometry and entity data.
func (rt *Runtime) handleGetEntryData(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return nil
	}
	return rt.sendEntryEntities(connID, userID, inst)
}

// handleMoveAvatar processes walk request.
func (rt *Runtime) handleMoveAvatar(connID string, userID int, body []byte) error {
	inst, entity := rt.findEntityByConnID(connID, userID)
	if inst == nil {
		return nil
	}
	r := packet.DecodeMoveAvatar(body)
	if r == nil {
		return nil
	}
	if rt.entitySvc != nil {
		_ = rt.entitySvc.Walk(context.Background(), inst, entity, r[0], r[1])
		return nil
	}
	reply := make(chan error, 1)
	inst.Send(engine.Message{Type: engine.MsgWalk, Entity: entity, TargetX: r[0], TargetY: r[1], Reply: reply})
	<-reply
	return nil
}

// sendRoomData transmits room loading geometry and settings.
func (rt *Runtime) sendRoomData(connID string, userID int, inst *engine.Instance, room domain.Room) error {
	if err := rt.sendPacket(connID, packet.RoomReadyComposer{
		ModelSlug: inst.Layout.Slug, RoomID: int32(room.ID),
	}); err != nil {
		return err
	}
	hm := heightmap.EncodeFloorMapWithDoor(inst.Layout.Grid, inst.Layout.DoorX, inst.Layout.DoorY, inst.Layout.DoorZ)
	if err := rt.sendPacket(connID, packet.FloorHeightMapComposer{
		Scale: true, WallHeight: int32(inst.Layout.WallHeight), Heightmap: hm,
	}); err != nil {
		return err
	}
	stacking := heightmap.EncodeStackingMap(inst.Layout.Grid)
	w := inst.Layout.Width()
	if err := rt.sendPacket(connID, packet.HeightMapComposer{
		Width: int32(w), TotalTiles: int32(len(stacking)), Heights: stacking,
	}); err != nil {
		return err
	}
	if err := rt.sendPacket(connID, packet.RoomEntryInfoComposer{
		RoomID: int32(room.ID), IsOwner: room.OwnerID == userID,
	}); err != nil {
		return err
	}
	if err := rt.sendRoomPermissions(connID, userID, room); err != nil {
		return err
	}
	if err := rt.sendPacket(connID, packet.RoomVisualizationComposer{
		WallThickness: int32(room.WallThickness), FloorThickness: int32(room.FloorThickness),
	}); err != nil {
		return err
	}
	if err := rt.sendPacket(connID, packet.FurnitureAliasesComposer{}); err != nil {
		return err
	}
	canVote := true
	if rt.voteRepo != nil {
		voted, _ := rt.voteRepo.HasVoted(context.Background(), room.ID, userID)
		canVote = !voted
	}
	if err := rt.sendPacket(connID, packet.RoomScoreComposer{
		Score: int32(room.Score), CanVote: canVote,
	}); err != nil {
		return err
	}
	if rt.floorItemSender != nil {
		return rt.floorItemSender(context.Background(), connID, inst.RoomID)
	}
	return nil
}

// sendRoomPermissions transmits the recipient owner or controller state for the active room.
func (rt *Runtime) sendRoomPermissions(connID string, userID int, room domain.Room) error {
	if room.OwnerID == userID {
		if err := rt.sendPacket(connID, packet.YouAreOwnerComposer{}); err != nil {
			return err
		}
		return rt.sendPacket(connID, packet.YouAreControllerComposer{Level: 4})
	}
	if rt.service.HasRights(context.Background(), room.ID, userID) {
		return rt.sendPacket(connID, packet.YouAreControllerComposer{Level: 1})
	}
	return rt.sendPacket(connID, packet.YouAreNotControllerComposer{})
}

// sendEntryEntities transmits entity list and enters the user.
// If the user already has an entity in the room (e.g. duplicate GetRoomEntryData), it skips creation.
func (rt *Runtime) sendEntryEntities(connID string, userID int, inst *engine.Instance) error {
	_, existing := rt.findEntityByConnID(connID, userID)
	if existing == nil {
		existingEntities := inst.Entities()
		username, look, motto, gender := "", "", "", "M"
		if rt.profileResolver != nil {
			if u, l, m, g, err := rt.profileResolver(context.Background(), userID); err == nil {
				username, look, motto, gender = u, l, m, g
			}
		}
		entity := domain.NewPlayerEntity(0, userID, connID, username, look, motto, gender,
			domain.Tile{X: inst.Layout.DoorX, Y: inst.Layout.DoorY, Z: inst.Layout.DoorZ, State: domain.TileOpen})
		if err := rt.service.EnterRoom(context.Background(), inst, &entity, inst.RoomID, userID); err != nil {
			return err
		}
		ctx := context.Background()
		newEntities := []domain.RoomEntity{entity}
		usersBody, usersErr := packet.UsersComposer{Entities: newEntities}.Encode()
		if usersErr == nil {
			frame := codec.EncodeFrame(packet.UsersComposerID, usersBody)
			for _, e := range existingEntities {
				if e.Type == domain.EntityPlayer && e.UserID != 0 && e.UserID != userID {
					_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(e.UserID), frame)
				}
			}
		}
		updateBody, updateErr := packet.UserUpdateComposer{Entities: newEntities}.Encode()
		if updateErr == nil {
			frame := codec.EncodeFrame(packet.UserUpdateComposerID, updateBody)
			for _, e := range existingEntities {
				if e.Type == domain.EntityPlayer && e.UserID != 0 && e.UserID != userID {
					_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(e.UserID), frame)
				}
			}
		}
	}
	entities := inst.Entities()
	if err := rt.sendPacket(connID, packet.UsersComposer{Entities: entities}); err != nil {
		return err
	}
	return rt.sendPacket(connID, packet.UserUpdateComposer{Entities: entities})
}
