package realtime

import (
	"context"
	"fmt"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
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
	case packet.SanctionTradeLockPacketID:
		return true, rt.handleTradeLock(ctx, connID, userID, body)
	case packet.CallForHelpPacketID:
		return true, rt.handleCallForHelp(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleModKick processes a moderator hotel kick.
func (rt *Runtime) handleModKick(ctx context.Context, _ string, issuerID int, body []byte) error {
	var pkt packet.ModKickUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeKick,
		TargetUserID: targetID, IssuerID: issuerID, Reason: pkt.Message,
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod kick failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, pkt.Message)
	rt.disconnectUser(ctx, targetID)
	rt.alertAmbassadors(ctx, fmt.Sprintf("User %d kicked by %d: %s", targetID, issuerID, pkt.Message))
	return nil
}

// handleModMute processes a moderator hotel mute.
func (rt *Runtime) handleModMute(ctx context.Context, _ string, issuerID int, body []byte) error {
	var pkt packet.ModMuteUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeMute,
		TargetUserID: targetID, IssuerID: issuerID,
		Reason: pkt.Message, DurationMinutes: int(pkt.Minutes),
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod mute failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, pkt.Message)
	return nil
}

// handleModBan processes a moderator hotel ban.
func (rt *Runtime) handleModBan(ctx context.Context, _ string, issuerID int, body []byte) error {
	var pkt packet.ModBanUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	durationMinutes := int(pkt.Duration) * 60
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeBan,
		TargetUserID: targetID, IssuerID: issuerID,
		Reason: pkt.Message, DurationMinutes: durationMinutes,
	}
	if durationMinutes > 0 {
		exp := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		action.ExpiresAt = &exp
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod ban failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendBanToUser(ctx, targetID, pkt.Message)
	rt.disconnectUser(ctx, targetID)
	rt.alertAmbassadors(ctx, fmt.Sprintf("User %d banned by %d: %s", targetID, issuerID, pkt.Message))
	return nil
}

// handleModWarn processes a moderator warning/caution.
func (rt *Runtime) handleModWarn(ctx context.Context, _ string, issuerID int, body []byte) error {
	var pkt packet.ModWarnUserPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeWarn,
		TargetUserID: targetID, IssuerID: issuerID, Reason: pkt.Message,
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("mod warn failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, pkt.Message)
	return nil
}

// sendCautionToUser sends a ModerationCaution packet to a target user.
func (rt *Runtime) sendCautionToUser(ctx context.Context, userID int, message string) {
	pkt := notificationpacket.ModerationCautionPacket{Message: message, Detail: ""}
	body, err := pkt.Encode()
	if err != nil {
		return
	}
	frame := codec.EncodeFrame(notificationpacket.ModerationCautionPacketID, body)
	_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), frame)
}

// sendBanToUser sends a UserBanned packet to a target user.
func (rt *Runtime) sendBanToUser(ctx context.Context, userID int, message string) {
	pkt := bannedpacket.UserBannedPacket{Message: message}
	body, err := pkt.Encode()
	if err != nil {
		return
	}
	frame := codec.EncodeFrame(bannedpacket.UserBannedPacketID, body)
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
			_ = rt.closer.Close(ctx, s.ConnID, 1000, "moderation")
		}
	}
}
