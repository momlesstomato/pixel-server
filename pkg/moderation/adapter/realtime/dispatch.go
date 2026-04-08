package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	bannedpacket "github.com/momlesstomato/pixel-server/pkg/user/packet/banned"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated moderation packet.
func (rt *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := rt.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.ModKickUserPacketID:
		return true, rt.handleModKick(ctx, connID, userID, body)
	case packet.ModMuteUserPacketID:
		return true, rt.handleModMute(ctx, connID, userID, body)
	case packet.ModBanUserPacketID:
		return true, rt.handleModBan(ctx, connID, userID, body)
	case packet.ModWarnUserPacketID:
		return true, rt.handleModWarn(ctx, connID, userID, body)
	case packet.ModAlertUserPacketID:
		return true, rt.handleModAlert(ctx, connID, userID, body)
	case packet.ModRoomAlertPacketID:
		return rt.handleModRoomAlertOrPass(ctx, connID, userID, body)
	case packet.RoomAmbassadorAlertPacketID:
		return true, rt.handleRoomAmbassadorAlert(ctx, connID, userID, body)
	case packet.ModToolRequestRoomInfoPacketID:
		return true, rt.handleModToolRequestRoomInfo(ctx, connID, userID, body)
	case packet.ModToolChangeRoomSettingsPacketID:
		return true, rt.handleModToolChangeRoomSettings(ctx, connID, userID, body)
	case packet.ModToolRequestRoomChatlogPacketID:
		return true, rt.handleModToolRequestRoomChatlog(ctx, connID, userID, body)
	case packet.ModToolUserInfoPacketID:
		return true, rt.handleModToolUserInfo(ctx, connID, userID, body)
	case packet.GetPendingCallsForHelpPacketID:
		return true, rt.handleGetPendingCallsForHelp(ctx, connID, userID)
	case packet.GetCFHChatlogPacketID:
		return true, rt.handleGetCFHChatlog(ctx, connID, userID, body)
	case packet.ModToolPreferencesPacketID:
		return true, rt.handleModToolPreferences(ctx, connID, userID)
	case packet.RoomMutePacketID:
		return true, rt.handleRoomMute(ctx, connID, userID)
	case packet.SanctionTradeLockPacketID:
		return true, rt.handleTradeLock(ctx, connID, userID, body)
	case packet.CallForHelpPacketID:
		return true, rt.handleCallForHelp(ctx, connID, userID, body)
	case packet.GetCFHStatusPacketID:
		return true, rt.handleGetCFHStatus(ctx, connID, userID)
	case packet.GuideSessionCreatePacketID:
		return true, rt.handleGuideSessionCreate(ctx, connID, userID, body)
	case packet.GetGuideReportingStatusPacketID:
		return true, rt.handleGetGuideReportingStatus(ctx, connID, userID)
	default:
		return false, nil
	}
}

// handleModKick processes a moderator hotel kick.
func (rt *Runtime) handleModKick(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermKick)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.ModKickUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeKick,
		TargetUserID: targetID, IssuerID: issuerID, Reason: rt.actionReason(ctx, issuerID, pkt.Message, "kick"),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod kick failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, action.Reason)
	rt.disconnectUser(ctx, targetID)
	rt.alertAmbassadors(ctx, fmt.Sprintf("User %d kicked by %d: %s", targetID, issuerID, action.Reason))
	return nil
}

// handleModMute processes a moderator hotel mute.
func (rt *Runtime) handleModMute(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermMute)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.ModMuteUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeMute,
		TargetUserID: targetID, IssuerID: issuerID,
		Reason: rt.actionReason(ctx, issuerID, pkt.Message, "mute"), DurationMinutes: int(pkt.Minutes),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod mute failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, action.Reason)
	return nil
}

// handleModBan processes a moderator hotel ban.
func (rt *Runtime) handleModBan(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermBan)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.ModBanUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	durationMinutes := int(pkt.Duration) * 60
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeBan,
		TargetUserID: targetID, IssuerID: issuerID,
		Reason: rt.actionReason(ctx, issuerID, pkt.Message, "ban"), DurationMinutes: durationMinutes,
	}
	if durationMinutes > 0 {
		exp := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		action.ExpiresAt = &exp
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod ban failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendBanToUser(ctx, targetID, action.Reason)
	rt.disconnectUser(ctx, targetID)
	rt.alertAmbassadors(ctx, fmt.Sprintf("User %d banned by %d: %s", targetID, issuerID, action.Reason))
	return nil
}

// handleModWarn processes a moderator warning/caution.
func (rt *Runtime) handleModWarn(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermWarn)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.ModWarnUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeWarn,
		TargetUserID: targetID, IssuerID: issuerID, Reason: rt.actionReason(ctx, issuerID, pkt.Message, "warning"),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod warn failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, action.Reason)
	return nil
}

// handleModAlert processes a moderator direct alert/message.
func (rt *Runtime) handleModAlert(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermWarn)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.ModAlertUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeWarn,
		TargetUserID: targetID, IssuerID: issuerID, Reason: rt.actionReason(ctx, issuerID, pkt.Message, "alert"),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod alert failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, action.Reason)
	return nil
}

// handleModRoomAlert processes a moderator current-room alert broadcast.
func (rt *Runtime) handleModRoomAlert(ctx context.Context, connID string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermWarn)
		if err != nil || !ok {
			return nil
		}
	}
	if rt.roomAlertSender == nil {
		return nil
	}
	var pkt packet.ModRoomAlertPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	message := rt.actionReason(ctx, issuerID, pkt.Message, "room alert")
	if rt.currentRoomID != nil {
		if roomID, ok := rt.currentRoomID(connID); ok {
			action := &domain.Action{
				Scope:      domain.ScopeRoom,
				ActionType: domain.TypeWarn,
				IssuerID:   issuerID,
				RoomID:     roomID,
				Reason:     message,
			}
			if err := rt.service.Create(ctx, action); err != nil {
				rt.logger.Warn("mod room alert registry failed", zap.Int("issuer", issuerID), zap.Int("room_id", roomID), zap.Error(err))
			}
		}
	}
	if err := rt.roomAlertSender(ctx, connID, message); err != nil {
		rt.logger.Warn("mod room alert failed", zap.Int("issuer", issuerID), zap.Error(err))
	}
	return nil
}

// handleModRoomAlertOrPass only claims packet 3842 when it decodes as a moderation room alert.
func (rt *Runtime) handleModRoomAlertOrPass(ctx context.Context, connID string, issuerID int, body []byte) (bool, error) {
	var pkt packet.ModRoomAlertPacket
	if err := pkt.Decode(body); err != nil {
		return false, nil
	}
	return true, rt.handleModRoomAlert(ctx, connID, issuerID, body)
}

// sendCautionToUser sends a ModerationCaution packet to a target user.
func (rt *Runtime) sendCautionToUser(ctx context.Context, userID int, message string) {
	rt.sendPacketToUser(ctx, userID, notificationpacket.ModerationCautionPacket{Message: message, Detail: ""})
}

// sendBanToUser sends a UserBanned packet to a target user.
func (rt *Runtime) sendBanToUser(ctx context.Context, userID int, message string) {
	rt.sendPacketToUser(ctx, userID, bannedpacket.UserBannedPacket{Message: message})
}

// sendPacketToUser delivers one packet directly when the target connection is local and falls back to pub/sub otherwise.
func (rt *Runtime) sendPacketToUser(ctx context.Context, userID int, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	if userID <= 0 {
		return
	}
	body, err := pkt.Encode()
	if err != nil {
		return
	}
	if session, ok := rt.sessions.FindByUserID(userID); ok {
		if err := rt.transport.Send(session.ConnID, pkt.PacketID(), body); err == nil {
			return
		}
	}
	frame := codec.EncodeFrame(pkt.PacketID(), body)
	_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), frame)
}

// disconnectUser closes the session for a target user.
func (rt *Runtime) disconnectUser(ctx context.Context, userID int) {
	if rt.closer == nil {
		return
	}
	sessions, err := rt.sessions.ListAll()
	if err != nil {
		return
	}
	for _, s := range sessions {
		if s.UserID == userID {
			if rt.roomLeaveNotifier != nil {
				rt.roomLeaveNotifier(s.ConnID)
			}
			_ = rt.closer.Close(ctx, s.ConnID, 1000, "moderation")
		}
	}
}

// actionReason returns the moderation reason or a deterministic role-aware fallback.
func (rt *Runtime) actionReason(ctx context.Context, issuerID int, message string, action string) string {
	if message != "" {
		return message
	}
	prefix := "moderator"
	if rt.permissions != nil {
		if ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermAmbassador); err == nil && ok {
			prefix = "ambassador"
		}
	}
	return prefix + " " + action
}
