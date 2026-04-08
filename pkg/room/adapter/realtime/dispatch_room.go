package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// triggerDoorbell sends a doorbell notification to the room owner and parks the visitor.
func (rt *Runtime) triggerDoorbell(ctx context.Context, connID string, userID int, username string, room domain.Room) error {
	rt.access.mu.Lock()
	rt.access.pendingDoorbell[username] = doorbellEntry{Username: username, UserID: userID, ConnID: connID, RoomID: room.ID}
	rt.access.mu.Unlock()
	inst, ok := rt.service.Manager().Get(room.ID)
	if !ok {
		rt.access.mu.Lock()
		delete(rt.access.pendingDoorbell, username)
		rt.access.mu.Unlock()
		return rt.sendNoRoomPresentFeedback(connID)
	}
	notified := false
	for _, entity := range inst.Entities() {
		if entity.Type != domain.EntityPlayer || entity.UserID == 0 {
			continue
		}
		if !rt.canControlDoorbell(ctx, room, entity.UserID) {
			continue
		}
		if err := rt.sendPacket(entity.ConnID, packet.DoorbellComposer{Username: username}); err != nil {
			rt.logger.Warn("doorbell notify failed", zap.String("conn_id", entity.ConnID), zap.Error(err))
			continue
		}
		notified = true
	}
	if notified {
		return nil
	}
	rt.access.mu.Lock()
	delete(rt.access.pendingDoorbell, username)
	rt.access.mu.Unlock()
	return rt.sendNoRoomPresentFeedback(connID)
}

// handleLetUserIn processes room owner doorbell approval or denial.
func (rt *Runtime) handleLetUserIn(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.LetUserInPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	rt.access.mu.Lock()
	entry, ok := rt.access.pendingDoorbell[pkt.Username]
	rt.access.mu.Unlock()
	if !ok {
		return nil
	}
	roomID, inRoom := rt.roomIDByConn(connID)
	if !inRoom || roomID != entry.RoomID {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, entry.RoomID)
	if err != nil {
		return rt.sendPacket(entry.ConnID, packet.CantConnectComposer{ErrorCode: 1})
	}
	if !rt.canControlDoorbell(ctx, room, userID) {
		return nil
	}
	rt.access.mu.Lock()
	delete(rt.access.pendingDoorbell, pkt.Username)
	rt.access.mu.Unlock()
	if !pkt.Let {
		return rt.sendPacket(entry.ConnID, packet.FlatAccessibleComposer{Username: pkt.Username, Accessible: false})
	}
	inst, ok := rt.service.Manager().Get(entry.RoomID)
	if !ok {
		return rt.sendPacket(entry.ConnID, packet.CantConnectComposer{ErrorCode: 1})
	}
	if err := rt.sendPacket(entry.ConnID, packet.FlatAccessibleComposer{Username: pkt.Username, Accessible: true}); err != nil {
		return err
	}
	if err := rt.sendPacket(entry.ConnID, packet.OpenConnectionComposer{}); err != nil {
		return err
	}
	rt.setRoomForConn(entry.ConnID, entry.RoomID)
	rt.clearPasswordFailures(entry.UserID, entry.RoomID)
	visitorID, visitorFound := rt.userID(entry.ConnID)
	if !visitorFound {
		return nil
	}
	if err := rt.sendRoomData(entry.ConnID, visitorID, inst, room); err != nil {
		return err
	}
	return nil
}

// handleGetRoomSettings sends room settings to the owner or a room-master override.
func (rt *Runtime) handleGetRoomSettings(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.GetRoomSettingsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, int(pkt.RoomID))
	if err != nil {
		return nil
	}
	if !rt.canManageRoom(ctx, userID, room) {
		return nil
	}
	return rt.sendPacket(connID, packet.RoomSettingsComposer{Room: room})
}

// handleSaveRoomSettings persists updated room settings from the owner or a room-master override.
func (rt *Runtime) handleSaveRoomSettings(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.SaveRoomSettingsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, int(pkt.RoomID))
	if err != nil {
		return nil
	}
	if !rt.canManageRoom(ctx, userID, room) {
		return nil
	}
	passwordHash := ""
	if packet.IntToAccessState(pkt.State) == domain.AccessPassword {
		passwordHash = room.Password
		if pkt.Password != "" {
			hash, hashErr := bcrypt.GenerateFromPassword([]byte(pkt.Password), bcrypt.MinCost)
			if hashErr != nil {
				rt.logger.Warn("hash room password failed", zap.Int("room_id", int(pkt.RoomID)), zap.Error(hashErr))
				return nil
			}
			passwordHash = string(hash)
		}
	}
	updated := domain.Room{
		Name: pkt.Name, Description: pkt.Description,
		State:          packet.IntToAccessState(pkt.State),
		Password:       passwordHash,
		CategoryID:     int(pkt.CategoryID),
		MaxUsers:       int(pkt.MaxUsers),
		Tags:           pkt.Tags,
		AllowPets:      pkt.AllowPets,
		AllowTrading:   pkt.AllowTrading,
		TradeMode:      int(pkt.TradeMode),
		WallThickness:  int(pkt.WallThickness),
		FloorThickness: int(pkt.FloorThickness),
		WallHeight:     room.WallHeight,
	}
	saveErr := error(nil)
	if room.OwnerID == userID {
		saveErr = rt.service.SaveSettings(ctx, int(pkt.RoomID), userID, updated)
	} else {
		updated.ID = int(pkt.RoomID)
		saveErr = rt.service.ModerateSettings(ctx, updated)
	}
	if saveErr != nil {
		rt.logger.Warn("save room settings failed", zap.Int("room_id", int(pkt.RoomID)), zap.Error(saveErr))
		return nil
	}
	rt.broadcastToRoom(int(pkt.RoomID), packet.RoomVisualizationComposer{
		WallsHidden:    pkt.HideWalls,
		WallThickness:  pkt.WallThickness,
		FloorThickness: pkt.FloorThickness,
	})
	rt.broadcastToRoom(int(pkt.RoomID), packet.RoomChatSettingsComposer{
		Mode:       pkt.ChatMode,
		Weight:     pkt.ChatWeight,
		Speed:      pkt.ChatSpeed,
		Distance:   pkt.ChatDistance,
		Protection: pkt.ChatProtection,
	})
	rt.broadcastToRoom(int(pkt.RoomID), packet.RoomSettingsUpdatedComposer{RoomID: pkt.RoomID})
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
	roomID, ok := rt.roomIDByConn(connID)
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
				rt.publishPacketToUser(context.Background(), e.UserID, sessionnavigation.DesktopViewResponsePacket{})
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

// handleUnbanOrPass handles room unban only when the payload clearly targets the current room.
func (rt *Runtime) handleUnbanOrPass(ctx context.Context, connID string, userID int, body []byte) (bool, error) {
	var pkt packet.UnbanUserPacket
	if err := pkt.Decode(body); err != nil {
		return false, nil
	}
	roomID, ok := rt.roomIDByConn(connID)
	if !ok || roomID != int(pkt.RoomID) {
		return false, nil
	}
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil || room.OwnerID != userID {
		return false, nil
	}
	return true, rt.handleUnbanUser(ctx, connID, userID, body)
}

// handleAssignRights grants room rights to one user.
func (rt *Runtime) handleAssignRights(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.AssignRightsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	if err := rt.service.GrantRights(ctx, roomID, userID, int(pkt.UserID)); err != nil {
		rt.logger.Warn("grant rights failed", zap.Int("room_id", roomID), zap.Error(err))
		return nil
	}
	entry := packet.RightsEntry{UserID: pkt.UserID, Username: rt.resolveUsername(ctx, int(pkt.UserID))}
	if err := rt.sendPacket(connID, packet.RoomRightsAddedComposer{RoomID: int32(roomID), Entry: entry}); err != nil {
		return err
	}
	if target, found := rt.sessions.FindByUserID(int(pkt.UserID)); found {
		if targetRoomID, inRoom := rt.roomIDByConn(target.ConnID); inRoom && targetRoomID == roomID {
			if err := rt.sendPacket(target.ConnID, packet.YouAreControllerComposer{Level: 1}); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleRemoveRights revokes room rights from one user.
func (rt *Runtime) handleRemoveRights(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.RemoveRightsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	if err := rt.service.RevokeRights(ctx, roomID, userID, int(pkt.UserID)); err != nil {
		rt.logger.Warn("revoke rights failed", zap.Int("room_id", roomID), zap.Error(err))
		return nil
	}
	if err := rt.sendPacket(connID, packet.RoomRightsRemovedComposer{RoomID: int32(roomID), UserID: pkt.UserID}); err != nil {
		return err
	}
	if target, found := rt.sessions.FindByUserID(int(pkt.UserID)); found {
		if targetRoomID, inRoom := rt.roomIDByConn(target.ConnID); inRoom && targetRoomID == roomID {
			if err := rt.sendPacket(target.ConnID, packet.YouAreNotControllerComposer{}); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleRemoveMyRights removes rights for the requesting user.
func (rt *Runtime) handleRemoveMyRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	_ = rt.service.RevokeRights(ctx, roomID, userID, userID)
	return rt.sendPacket(connID, packet.YouAreNotControllerComposer{})
}

// handleRemoveAllRights removes all rights holders from the current room.
func (rt *Runtime) handleRemoveAllRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	holders, err := rt.service.ListRights(ctx, roomID, userID)
	if err != nil {
		return nil
	}
	if err := rt.service.RevokeAllRights(ctx, roomID, userID); err != nil {
		rt.logger.Warn("revoke all rights failed", zap.Int("room_id", roomID), zap.Error(err))
		return nil
	}
	for _, holderID := range holders {
		if err := rt.sendPacket(connID, packet.RoomRightsRemovedComposer{RoomID: int32(roomID), UserID: int32(holderID)}); err != nil {
			return err
		}
		if target, found := rt.sessions.FindByUserID(holderID); found {
			if targetRoomID, inRoom := rt.roomIDByConn(target.ConnID); inRoom && targetRoomID == roomID {
				if err := rt.sendPacket(target.ConnID, packet.YouAreNotControllerComposer{}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// handleGetRoomRights sends the rights holder list to the owner.
func (rt *Runtime) handleGetRoomRights(ctx context.Context, connID string, userID int) error {
	roomID, ok := rt.roomIDByConn(connID)
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
	roomID, ok := rt.roomIDByConn(connID)
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

// handleRoomMuteUser applies or clears one room-scoped mute for a target user.
func (rt *Runtime) handleRoomMuteUser(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.RoomMuteUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.roomIDByConn(connID)
	if !ok {
		return nil
	}
	if pkt.RoomID != 0 && int(pkt.RoomID) != roomID {
		return nil
	}
	room, err := rt.service.FindRoom(ctx, roomID)
	if err != nil {
		return nil
	}
	modMute := false
	if rt.permissions != nil {
		modMute, _ = rt.permissions.HasPermission(ctx, userID, "moderation.mute")
	}
	if room.OwnerID != userID && !rt.service.HasRights(ctx, roomID, userID) && !modMute {
		return nil
	}
	targetID := int(pkt.UserID)
	target, found := rt.sessions.FindByUserID(targetID)
	if !found {
		return nil
	}
	if targetRoomID, inRoom := rt.roomIDByConn(target.ConnID); !inRoom || targetRoomID != roomID {
		return nil
	}
	duration := time.Duration(pkt.Minutes) * time.Minute
	rt.setRoomUserMute(roomID, targetID, duration)
	message := "You are muted in this room right now."
	if pkt.Minutes > 0 {
		message = fmt.Sprintf("You are muted in this room for %d minute(s).", pkt.Minutes)
	}
	rt.publishPacketToUser(ctx, targetID, notificationpacket.GenericAlertPacket{Message: message})
	return nil
}

// handleKickUser removes a target user from the room on behalf of the room owner, rights holder, or moderator.
func (rt *Runtime) handleKickUser(ctx context.Context, connID string, userID int, body []byte) error {
	r := codec.NewReader(body)
	targetID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	roomID, ok := rt.roomIDByConn(connID)
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
	rt.publishPacketToUser(ctx, userID, notificationpacket.GenericErrorPacket{ErrorCode: 4008})
	rt.publishPacketToUser(ctx, userID, sessionnavigation.DesktopViewResponsePacket{})
}

// handleBanUser bans a target user from the room on behalf of the owner, rights holder, or moderator.
func (rt *Runtime) handleBanUser(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.BanUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	roomID, ok := rt.roomIDByConn(connID)
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
