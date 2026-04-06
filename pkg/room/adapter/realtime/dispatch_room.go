package realtime

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
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

// handleGiveRoomScore processes a room vote request.
func (rt *Runtime) handleGiveRoomScore(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.GiveRoomScorePacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.connRooms[connID]
	if !ok || rt.voteRepo == nil {
		return nil
	}
	voted, _ := rt.voteRepo.HasVoted(ctx, roomID, userID)
	if voted || pkt.Score < 1 {
		return rt.sendPacket(connID, packet.RoomScoreComposer{Score: 0, CanVote: false})
	}
	if err := rt.voteRepo.CastVote(ctx, roomID, userID); err != nil {
		rt.logger.Warn("room vote failed", zap.Int("room_id", roomID), zap.Error(err))
		return nil
	}
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil {
		return nil
	}
	return rt.sendPacket(connID, packet.RoomScoreComposer{Score: int32(room.Score), CanVote: false})
}

// handleDeleteRoom processes a room deletion request from the owner.
func (rt *Runtime) handleDeleteRoom(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.DeleteRoomPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID := int(pkt.RoomID)
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil || room.OwnerID != userID {
		return nil
	}
	if err := rt.service.SoftDelete(ctx, roomID); err != nil {
		rt.logger.Warn("room delete failed", zap.Int("room_id", roomID), zap.Error(err))
		return nil
	}
	inst, ok := rt.service.Manager().Get(roomID)
	if ok {
		for _, e := range inst.Entities() {
			if e.Type == domain.EntityPlayer && e.UserID != 0 {
				_ = rt.sendPacket(e.ConnID, packet.DesktopViewComposer{})
			}
		}
		rt.service.Manager().Unload(roomID)
	}
	return nil
}

// handleGetBannedUsers sends the ban list for a room to the owner.
func (rt *Runtime) handleGetBannedUsers(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.GetBannedUsersPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID := int(pkt.RoomID)
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil || room.OwnerID != userID {
		return nil
	}
	bans, err := rt.service.ListBans(ctx, roomID)
	if err != nil {
		return nil
	}
	entries := make([]packet.BannedUserEntry, len(bans))
	for i, ban := range bans {
		name := rt.resolveUsername(ctx, ban.UserID)
		entries[i] = packet.BannedUserEntry{UserID: int32(ban.UserID), Username: name}
	}
	return rt.sendPacket(connID, packet.BannedUsersComposer{RoomID: pkt.RoomID, Bans: entries})
}

// handleUnbanUser removes a ban entry for a user in a room.
func (rt *Runtime) handleUnbanUser(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.UnbanUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID := int(pkt.RoomID)
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil || room.OwnerID != userID {
		return nil
	}
	ban, err := rt.service.FindBan(ctx, roomID, int(pkt.UserID))
	if err != nil || ban == nil {
		return nil
	}
	_ = rt.service.RemoveBan(ctx, ban.ID)
	return nil
}

// handleAssignRights grants room rights to one user.
func (rt *Runtime) handleAssignRights(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.AssignRightsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	if err := rt.service.GrantRights(ctx, roomID, userID, int(pkt.UserID)); err != nil {
		rt.logger.Warn("grant rights failed", zap.Int("room_id", roomID), zap.Error(err))
	}
	return nil
}

// handleRemoveRights revokes room rights from one user.
func (rt *Runtime) handleRemoveRights(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.RemoveRightsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	if err := rt.service.RevokeRights(ctx, roomID, userID, int(pkt.UserID)); err != nil {
		rt.logger.Warn("revoke rights failed", zap.Int("room_id", roomID), zap.Error(err))
	}
	return nil
}

// handleRemoveMyRights removes rights for the requesting user.
func (rt *Runtime) handleRemoveMyRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	_ = rt.service.RevokeRights(ctx, roomID, userID, userID)
	return nil
}

// handleRemoveAllRights removes all rights holders from the current room.
func (rt *Runtime) handleRemoveAllRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	if err := rt.service.RevokeAllRights(ctx, roomID, userID); err != nil {
		rt.logger.Warn("revoke all rights failed", zap.Int("room_id", roomID), zap.Error(err))
	}
	return nil
}

// handleGetRoomRights sends the rights holder list to the owner.
func (rt *Runtime) handleGetRoomRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	holders, err := rt.service.ListRights(ctx, roomID, userID)
	if err != nil {
		return nil
	}
	entries := make([]packet.RightsEntry, len(holders))
	for i, uid := range holders {
		entries[i] = packet.RightsEntry{UserID: int32(uid), Username: rt.resolveUsername(ctx, uid)}
	}
	return rt.sendPacket(connID, packet.RoomRightsListComposer{RoomID: int32(roomID), Entries: entries})
}

// handleToggleMuteTool toggles the room global chat mute state.
func (rt *Runtime) handleToggleMuteTool(connID string, userID int) error {
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	room, err := rt.service.FindRoom(context.Background(), roomID)
	if err != nil || room.OwnerID != userID {
		return nil
	}
	inst, ok := rt.service.Manager().Get(roomID)
	if !ok {
		return nil
	}
	inst.SetMuted(!inst.Muted())
	return nil
}

// handleKickUser removes a target user from the room on behalf of the room owner, rights holder, or moderator.
func (rt *Runtime) handleKickUser(ctx context.Context, connID string, userID int, body []byte) error {
	r := codec.NewReader(body)
	targetID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil {
		return nil
	}
	modKick := false
	if rt.permissions != nil {
		modKick, _ = rt.permissions.HasPermission(ctx, userID, "moderation.kick")
	}
	if room.OwnerID != userID && !rt.service.HasRights(ctx, roomID, userID) && !modKick {
		return nil
	}
	target, found := rt.sessions.FindByUserID(int(targetID))
	if !found {
		return nil
	}
	rt.sendKickedPacket(ctx, int(targetID))
	rt.leaveCurrentRoom(target.ConnID)
	return nil
}

// sendKickedPacket notifies one user that they have been kicked out of the room.
func (rt *Runtime) sendKickedPacket(ctx context.Context, userID int) {
	pkt := notificationpacket.GenericErrorPacket{ErrorCode: 4008}
	body, err := pkt.Encode()
	if err != nil {
		return
	}
	frame := codec.EncodeFrame(notificationpacket.GenericErrorPacketID, body)
	_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), frame)
}

// handleBanUser bans a target user from the room on behalf of the owner, rights holder, or moderator.
func (rt *Runtime) handleBanUser(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.BanUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.connRooms[connID]
	if !ok {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil {
		return nil
	}
	modKick := false
	if rt.permissions != nil {
		modKick, _ = rt.permissions.HasPermission(ctx, userID, "moderation.kick")
	}
	if room.OwnerID != userID && !rt.service.HasRights(ctx, roomID, userID) && !modKick {
		return nil
	}
	targetID := int(pkt.UserID)
	var expiresAt *time.Time
	switch pkt.BanType {
	case "RWUAM_BAN_USER_HOUR":
		t := time.Now().Add(time.Hour)
		expiresAt = &t
	case "RWUAM_BAN_USER_DAY":
		t := time.Now().Add(24 * time.Hour)
		expiresAt = &t
	}
	ban := domain.RoomBan{RoomID: roomID, UserID: targetID, ExpiresAt: expiresAt}
	if _, err := rt.service.CreateBan(ctx, ban); err != nil {
		rt.logger.Warn("create room ban failed", zap.Int("room_id", roomID), zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	target, found := rt.sessions.FindByUserID(targetID)
	if !found {
		return nil
	}
	rt.sendKickedPacket(ctx, targetID)
	rt.leaveCurrentRoom(target.ConnID)
	return nil
}
