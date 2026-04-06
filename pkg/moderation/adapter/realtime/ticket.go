package realtime

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
	"go.uber.org/zap"
)

// handleCallForHelp processes a user's call-for-help submission.
func (rt *Runtime) handleCallForHelp(ctx context.Context, _ string, userID int, body []byte) error {
	if rt.tickets == nil {
		return nil
	}
	var pkt packet.CallForHelpPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	ticket := &domain.Ticket{
		ReporterID: userID, ReportedID: int(pkt.ReportedID), RoomID: int(pkt.RoomID),
		Category: fmt.Sprintf("%d", pkt.Category), Message: pkt.Message,
	}
	if err := rt.tickets.Submit(ctx, ticket); err != nil {
		rt.logger.Warn("cfh submit failed", zap.Int("reporter", userID), zap.Error(err))
		return nil
	}
	rt.service.AlertAmbassadors(ctx, fmt.Sprintf("CFH from user %d in room %d", userID, pkt.RoomID))
	return nil
}

// handleTradeLock processes a moderator trade lock sanction.
func (rt *Runtime) handleTradeLock(ctx context.Context, _ string, issuerID int, body []byte) error {
	if rt.permissions != nil {
		ok, err := rt.permissions.HasPermission(ctx, issuerID, domain.PermTradeLock)
		if err != nil || !ok {
			return nil
		}
	}
	var pkt packet.SanctionTradeLockPacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	targetID := int(pkt.UserID)
	durationMinutes := int(pkt.Duration) * 60
	action := &domain.Action{
		Scope: domain.ScopeHotel, ActionType: domain.TypeTradeLock,
		TargetUserID: targetID, IssuerID: issuerID,
		Reason: pkt.Message, DurationMinutes: durationMinutes,
	}
	if err := rt.service.Create(ctx, action); err != nil {
		rt.logger.Warn("trade lock failed", zap.Int("target", targetID), zap.Error(err))
		return nil
	}
	rt.sendCautionToUser(ctx, targetID, pkt.Message)
	return nil
}
