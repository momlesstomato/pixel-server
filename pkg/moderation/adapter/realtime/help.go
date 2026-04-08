package realtime

import (
	"context"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/packet"
)

const (
	guideReportingStatusOK            int32 = 0
	guideReportingStatusPendingTicket int32 = 1
	guideSessionErrorGeneric          int32 = 0
	guideSessionErrorNoHelpers        int32 = 1
)

// handleGetCFHStatus responds with the caller's sanction status for help flows.
func (rt *Runtime) handleGetCFHStatus(ctx context.Context, connID string, userID int) error {
	now := time.Now()
	actions, err := rt.service.List(ctx, domain.ListFilter{TargetUserID: userID, Limit: 50})
	if err != nil {
		return nil
	}
	isMuted, _ := rt.service.IsHotelMuted(ctx, userID)
	isTradeLocked, _ := rt.service.IsTradeLocked(ctx, userID)
	current := currentSanctionAction(actions, now)
	pkt := packet.CFHSanctionStatusPacket{
		SanctionName:         "ALERT",
		SanctionReason:       "cfh.reason.EMPTY",
		SanctionCreationTime: now.UTC().Format(time.RFC3339),
		NextSanctionName:     "ALERT",
		HasCustomMute:        isMuted,
	}
	if current != nil {
		pkt.IsSanctionNew = now.Sub(current.CreatedAt) < 24*time.Hour
		pkt.IsSanctionActive = actionIsActive(*current, now)
		pkt.SanctionName = sanctionName(current.ActionType)
		pkt.SanctionLengthHours = durationHours(current.DurationMinutes)
		if reason := strings.TrimSpace(current.Reason); reason != "" {
			pkt.SanctionReason = reason
		}
		if !current.CreatedAt.IsZero() {
			pkt.SanctionCreationTime = current.CreatedAt.UTC().Format(time.RFC3339)
		}
		pkt.ProbationHoursLeft = remainingHours(current.ExpiresAt, now)
		pkt.NextSanctionName, pkt.NextSanctionLengthHours = nextSanction(current.ActionType)
	}
	if isTradeLocked {
		pkt.IsSanctionActive = true
		if tradeLock := latestActionOfType(actions, domain.TypeTradeLock, now); tradeLock != nil && tradeLock.ExpiresAt != nil {
			pkt.TradeLockExpiryTime = tradeLock.ExpiresAt.UTC().Format(time.RFC3339)
		}
	}
	return rt.sendPacketToConn(connID, pkt)
}

// handleGuideSessionCreate returns a deterministic guide-session error when no helper flow is available.
func (rt *Runtime) handleGuideSessionCreate(ctx context.Context, connID string, userID int, body []byte) error {
	var pkt packet.GuideSessionCreatePacket
	if err := pkt.Decode(body); err != nil {
		return nil
	}
	errorCode := guideSessionErrorNoHelpers
	if ticket := rt.findReporterTicket(ctx, userID); ticket != nil {
		errorCode = guideSessionErrorGeneric
	}
	return rt.sendPacketToConn(connID, packet.GuideSessionErrorPacket{ErrorCode: errorCode})
}

// handleGetGuideReportingStatus returns room-report/guide pending status for the caller.
func (rt *Runtime) handleGetGuideReportingStatus(ctx context.Context, connID string, userID int) error {
	pkt := packet.GuideReportingStatusPacket{StatusCode: guideReportingStatusOK}
	if ticket := rt.findReporterTicket(ctx, userID); ticket != nil {
		pkt.StatusCode = guideReportingStatusPendingTicket
		pkt.PendingTicket = rt.pendingGuideTicket(ctx, *ticket)
	}
	return rt.sendPacketToConn(connID, pkt)
}

// findReporterTicket resolves the caller's currently open or in-progress help ticket.
func (rt *Runtime) findReporterTicket(ctx context.Context, userID int) *domain.Ticket {
	if rt.tickets == nil || userID <= 0 {
		return nil
	}
	for _, status := range []domain.TicketStatus{domain.TicketInProgress, domain.TicketOpen} {
		tickets, err := rt.tickets.List(ctx, status, 50)
		if err != nil {
			return nil
		}
		for i := range tickets {
			if tickets[i].ReporterID == userID {
				return &tickets[i]
			}
		}
	}
	return nil

}

// pendingGuideTicket builds the Nitro pending-ticket payload from one moderation ticket.
func (rt *Runtime) pendingGuideTicket(ctx context.Context, ticket domain.Ticket) packet.PendingGuideTicketData {
	data := packet.PendingGuideTicketData{
		Type:        1,
		SecondsAgo:  int32(time.Since(ticket.CreatedAt).Seconds()),
		Description: ticket.Message,
	}
	if data.SecondsAgo < 0 {
		data.SecondsAgo = 0
	}
	if rt.users != nil && ticket.ReportedID > 0 {
		user, err := rt.users.FindByID(ctx, ticket.ReportedID)
		if err == nil {
			data.OtherPartyName = user.Username
			data.OtherPartyFigure = user.Figure
		}
	}
	if rt.rooms != nil && ticket.RoomID > 0 {
		room, err := rt.rooms.FindRoom(ctx, ticket.RoomID)
		if err == nil {
			data.RoomName = room.Name
		}
	}
	return data
}

// currentSanctionAction selects the most relevant sanction for the caller status view.
func currentSanctionAction(actions []domain.Action, now time.Time) *domain.Action {
	for i := range actions {
		action := &actions[i]
		if !actionIsActive(*action, now) {
			continue
		}
		if action.ActionType == domain.TypeBan || action.ActionType == domain.TypeMute {
			return action
		}
	}
	for i := range actions {
		action := &actions[i]
		if action.ActionType == domain.TypeWarn || action.ActionType == domain.TypeBan || action.ActionType == domain.TypeMute {
			return action
		}
	}
	return nil
}

// latestActionOfType resolves the most recent active action of one type.
func latestActionOfType(actions []domain.Action, actionType domain.ActionType, now time.Time) *domain.Action {
	for i := range actions {
		action := &actions[i]
		if action.ActionType == actionType && actionIsActive(*action, now) {
			return action
		}
	}
	return nil
}

// actionIsActive returns whether one moderation action is still active.
func actionIsActive(action domain.Action, now time.Time) bool {
	return action.Active && (action.ExpiresAt == nil || action.ExpiresAt.After(now))
}

// sanctionName maps one moderation action type to the Nitro sanction label.
func sanctionName(actionType domain.ActionType) string {
	switch actionType {
	case domain.TypeMute:
		return "MUTE"
	case domain.TypeBan:
		return "BAN"
	default:
		return "ALERT"
	}
}

// nextSanction resolves the next escalation label and duration in hours.
func nextSanction(actionType domain.ActionType) (string, int32) {
	switch actionType {
	case domain.TypeWarn:
		return "MUTE", 2
	case domain.TypeMute:
		return "BAN", 24
	default:
		return "BAN", 0
	}
}

// durationHours converts moderation duration minutes to whole hours rounded up.
func durationHours(durationMinutes int) int32 {
	if durationMinutes <= 0 {
		return 0
	}
	return int32((durationMinutes + 59) / 60)
}

// remainingHours returns remaining active duration in whole hours rounded up.
func remainingHours(expiresAt *time.Time, now time.Time) int32 {
	if expiresAt == nil || !expiresAt.After(now) {
		return 0
	}
	minutes := int(expiresAt.Sub(now).Minutes())
	return durationHours(minutes)
}
