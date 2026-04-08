package realtime

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	roomdomain "github.com/momlesstomato/pixel-server/pkg/room/domain"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	"go.uber.org/zap"
)

const moderatedRoomTitle = "Inappropriate to hotel staff"

func (rt *Runtime) handleRoomAmbassadorAlert(ctx context.Context, _ string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermWarn, domain.PermAmbassador) {
		return nil
	}
	var pkt packet.RoomAmbassadorAlertPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope:        domain.ScopeHotel,
		ActionType:   domain.TypeWarn,
		TargetUserID: targetID,
		IssuerID:     issuerID,
		Reason:       rt.actionReason(ctx, issuerID, "", "ambassador alert"),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("ambassador alert failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, action.Reason)
	return nil
}

func (rt *Runtime) handleModToolRequestRoomInfo(ctx context.Context, connID string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool, domain.PermHistory) || rt.rooms == nil {
		return nil
	}
	var pkt packet.ModToolRequestRoomInfoPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	room, err := rt.rooms.FindRoom(ctx, int(pkt.RoomID))
	if err != nil {
		return nil
	}
	userCount := 0
	if rt.roomUserCount != nil {
		userCount = rt.roomUserCount(room.ID)
	}
	ownerInRoom := false
	if rt.currentRoomID != nil {
		if session, ok := rt.sessions.FindByUserID(room.OwnerID); ok {
			if roomID, inRoom := rt.currentRoomID(session.ConnID); inRoom && roomID == room.ID {
				ownerInRoom = true
			}
		}
	}
	return rt.sendPacketToConn(connID, packet.ModToolRoomInfoPacket{
		RoomID:      int32(room.ID),
		UserCount:   int32(userCount),
		OwnerInRoom: ownerInRoom,
		OwnerID:     int32(room.OwnerID),
		OwnerName:   room.OwnerName,
		Exists:      true,
		Name:        room.Name,
		Description: room.Description,
		Tags:        room.Tags,
	})
}

func (rt *Runtime) handleModToolRequestRoomChatlog(ctx context.Context, connID string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool, domain.PermHistory) || rt.rooms == nil || rt.chatLogs == nil {
		return nil
	}
	var pkt packet.ModToolRequestRoomChatlogPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	room, err := rt.rooms.FindRoom(ctx, int(pkt.RoomID))
	if err != nil {
		return nil
	}
	entries, err := rt.chatLogs.ListByRoom(ctx, room.ID, time.Now().Add(-30*24*time.Hour), time.Now().Add(time.Minute))
	if err != nil {
		return nil
	}
	lines := make([]packet.RoomChatlogLine, len(entries))
	for i, entry := range entries {
		lines[i] = packet.RoomChatlogLine{
			Timestamp: entry.CreatedAt.Local().Format("15:04"),
			UserID:    int32(entry.UserID),
			Username:  entry.Username,
			Message:   entry.Message,
		}
	}
	return rt.sendPacketToConn(connID, packet.ModToolRoomChatlogPacket{RoomID: int32(room.ID), RoomName: room.Name, Chatlog: lines})
}

func (rt *Runtime) handleModToolUserInfo(ctx context.Context, connID string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool, domain.PermHistory) || rt.users == nil {
		return nil
	}
	var pkt packet.ModToolUserInfoPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	user, err := rt.users.FindByID(ctx, int(pkt.UserID))
	if err != nil {
		return nil
	}
	actions, err := rt.service.List(ctx, domain.ListFilter{TargetUserID: user.ID, Limit: 100})
	if err != nil {
		return nil
	}
	cautionCount := 0
	banCount := 0
	tradeLockCount := 0
	lastSanctionTime := ""
	lastSanctionAgeHours := int32(0)
	for _, action := range actions {
		switch action.ActionType {
		case domain.TypeWarn:
			cautionCount++
		case domain.TypeBan:
			banCount++
		case domain.TypeTradeLock:
			tradeLockCount++
		}
		if lastSanctionTime == "" && !action.CreatedAt.IsZero() {
			lastSanctionTime = action.CreatedAt.UTC().Format(time.RFC3339)
			lastSanctionAgeHours = int32(time.Since(action.CreatedAt).Hours())
		}
	}
	_, online := rt.sessions.FindByUserID(user.ID)
	return rt.sendPacketToConn(connID, packet.ModeratorUserInfoPacket{
		UserID:                  int32(user.ID),
		Username:                user.Username,
		Figure:                  user.Figure,
		RegistrationAgeMinutes:  0,
		MinutesSinceLastLogin:   0,
		Online:                  online,
		CFHCount:                0,
		AbusiveCFHCount:         0,
		CautionCount:            int32(cautionCount),
		BanCount:                int32(banCount),
		TradingLockCount:        int32(tradeLockCount),
		TradingExpiryDate:       "",
		LastPurchaseDate:        "",
		IdentityID:              int32(user.ID),
		IdentityRelatedBanCount: 0,
		PrimaryEmailAddress:     "",
		UserClassification:      fmt.Sprintf("group:%d", user.GroupID),
		LastSanctionTime:        lastSanctionTime,
		SanctionAgeHours:        lastSanctionAgeHours,
	})
}

func (rt *Runtime) handleGetPendingCallsForHelp(ctx context.Context, connID string, issuerID int) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool) || rt.tickets == nil {
		return nil
	}
	tickets, err := rt.tickets.List(ctx, domain.TicketOpen, 50)
	if err != nil {
		return nil
	}
	entries := make([]packet.CFHPendingEntry, len(tickets))
	for i, ticket := range tickets {
		entries[i] = packet.CFHPendingEntry{
			CallID:    strconv.FormatInt(ticket.ID, 10),
			Timestamp: ticket.CreatedAt.Local().Format("15:04"),
			Message:   ticket.Message,
		}
	}
	return rt.sendPacketToConn(connID, packet.CFHPendingPacket{Entries: entries})
}

func (rt *Runtime) handleGetCFHChatlog(ctx context.Context, connID string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool, domain.PermHistory) || rt.tickets == nil || rt.rooms == nil || rt.chatLogs == nil {
		return nil
	}
	var pkt packet.GetCFHChatlogPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	ticket, err := rt.tickets.FindByID(ctx, int64(pkt.TicketID))
	if err != nil {
		return nil
	}
	room, err := rt.rooms.FindRoom(ctx, ticket.RoomID)
	if err != nil {
		return nil
	}
	entries, err := rt.chatLogs.ListByRoom(ctx, room.ID, time.Now().Add(-30*24*time.Hour), time.Now().Add(time.Minute))
	if err != nil {
		return nil
	}
	lines := make([]packet.RoomChatlogLine, len(entries))
	for i, entry := range entries {
		lines[i] = packet.RoomChatlogLine{
			Timestamp: entry.CreatedAt.Local().Format("15:04"),
			UserID:    int32(entry.UserID),
			Username:  entry.Username,
			Message:   entry.Message,
		}
	}
	return rt.sendPacketToConn(connID, packet.ModeratorCFHChatlogPacket{
		TicketID:     int32(ticket.ID),
		ReporterID:   int32(ticket.ReporterID),
		ReportedID:   int32(ticket.ReportedID),
		ChatRecordID: int32(ticket.ID),
		RoomID:       int32(room.ID),
		RoomName:     room.Name,
		Chatlog:      lines,
	})
}

func (rt *Runtime) handleModToolPreferences(ctx context.Context, connID string, issuerID int) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool) {
		return nil
	}
	return rt.sendPacketToConn(connID, packet.ModeratorToolPreferencesPacket{WindowWidth: 275, WindowHeight: 400})
}

func (rt *Runtime) handleRoomMute(ctx context.Context, connID string, issuerID int) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermMute) || rt.currentRoomID == nil || rt.roomMuteToggler == nil {
		return nil
	}
	roomID, ok := rt.currentRoomID(connID)
	if !ok {
		return nil
	}
	if _, err := rt.roomMuteToggler(ctx, roomID); err != nil {
		rt.logger.Warn("room mute toggle failed", zap.Int("room_id", roomID), zap.Error(err))
	}
	return nil
}

func (rt *Runtime) handleModToolChangeRoomSettings(ctx context.Context, _ string, issuerID int, body []byte) error {
	if !rt.hasAnyPermission(ctx, issuerID, domain.PermTool) {
		return nil
	}
	var pkt packet.ModToolChangeRoomSettingsPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	if (pkt.LockDoor == 1 || pkt.ChangeTitle == 1) && rt.rooms != nil && rt.roomSettingsUpdater != nil {
		room, err := rt.rooms.FindRoom(ctx, int(pkt.RoomID))
		if err != nil {
			rt.logger.Warn("moderator room settings lookup failed", zap.Int("room_id", int(pkt.RoomID)), zap.Error(err))
		} else {
			if pkt.LockDoor == 1 {
				room.State = roomdomain.AccessLocked
			}
			if pkt.ChangeTitle == 1 {
				room.Name = moderatedRoomTitle
			}
			if err := rt.roomSettingsUpdater(ctx, room); err != nil {
				rt.logger.Warn("moderator room settings update failed", zap.Int("room_id", room.ID), zap.Error(err))
			}
		}
	}
	if pkt.KickUsers == 1 {
		rt.kickUsersFromRoom(ctx, int(pkt.RoomID), issuerID)
	}
	return nil
}

func (rt *Runtime) kickUsersFromRoom(ctx context.Context, roomID int, issuerID int) {
	if roomID <= 0 || rt.currentRoomID == nil || rt.roomLeaveNotifier == nil {
		return
	}
	sessions, err := rt.sessions.ListAll()
	if err != nil {
		return
	}
	for _, session := range sessions {
		if session.UserID == issuerID {
			continue
		}
		currentRoomID, ok := rt.currentRoomID(session.ConnID)
		if !ok || currentRoomID != roomID {
			continue
		}
		rt.sendPacketToUser(ctx, session.UserID, notificationpacket.GenericErrorPacket{ErrorCode: 4008})
		rt.sendPacketToUser(ctx, session.UserID, sessionnavigation.DesktopViewResponsePacket{})
		rt.roomLeaveNotifier(session.ConnID)
	}
}

func (rt *Runtime) sendPacketToConn(connID string, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) error {
	body, err := pkt.Encode()
	if err != nil {
		return err
	}
	return rt.transport.Send(connID, pkt.PacketID(), body)
}

func (rt *Runtime) hasAnyPermission(ctx context.Context, userID int, scopes ...string) bool {
	if rt.permissions == nil {
		return true
	}
	for _, scope := range scopes {
		ok, err := rt.permissions.HasPermission(ctx, userID, scope)
		if err == nil && ok {
			return true
		}
	}
	return false
}
