package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdkmessenger "github.com/momlesstomato/pixel-sdk/events/messenger"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

const maxMessageLength = 255

// SendMessage routes a private message from sender to recipient.
// Routes online messages via broadcaster; stores offline messages in the DB.
func (service *Service) SendMessage(ctx context.Context, connID string, fromUserID, toUserID int, message string) error {
	if fromUserID <= 0 || toUserID <= 0 {
		return fmt.Errorf("user ids must be positive")
	}
	message = strings.TrimSpace(message)
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength]
	}
	if err := service.checkFlood(ctx, connID, fromUserID); err != nil {
		return err
	}
	if service.fire != nil {
		event := &sdkmessenger.PrivateMessageSent{ConnID: connID, FromUserID: fromUserID, ToUserID: toUserID, Message: message}
		service.fire(event)
		if event.Cancelled() {
			return domain.ErrSenderMuted
		}
	}
	friends, err := service.repository.AreFriends(ctx, fromUserID, toUserID)
	if err != nil {
		return err
	}
	if !friends {
		return domain.ErrNotFriends
	}
	_, online := service.sessions.FindByUserID(toUserID)
	if !online {
		return service.repository.SaveOfflineMessage(ctx, fromUserID, toUserID, message)
	}
	return nil
}

// DeliverOfflineMessages fetches and removes offline messages for one user.
func (service *Service) DeliverOfflineMessages(ctx context.Context, userID int) ([]domain.OfflineMessage, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.GetAndDeleteOfflineMessages(ctx, userID)
}

// SendRoomInvite routes a room invite from sender to recipients.
func (service *Service) SendRoomInvite(ctx context.Context, connID string, fromUserID int, toUserIDs []int, message string) error {
	if fromUserID <= 0 {
		return fmt.Errorf("from user id must be positive")
	}
	message = strings.TrimSpace(message)
	if len(message) > maxMessageLength {
		message = message[:maxMessageLength]
	}
	if service.fire != nil {
		event := &sdkmessenger.RoomInviteSent{ConnID: connID, FromUserID: fromUserID, ToUserIDs: toUserIDs, Message: message}
		service.fire(event)
		if event.Cancelled() {
			return domain.ErrSenderMuted
		}
	}
	return nil
}

// PurgeOldOfflineMessages deletes messages older than the configured TTL.
func (service *Service) PurgeOldOfflineMessages(ctx context.Context) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -service.config.OfflineMsgTTLDays)
	return service.repository.DeleteOfflineMessagesOlderThan(ctx, cutoff.Unix())
}

// checkFlood enforces message rate limiting for one connection.
// When the sender holds messenger.flood.bypass the check is skipped entirely.
func (service *Service) checkFlood(ctx context.Context, connID string, userID int) error {
	if service.checker != nil {
		if ok, _ := service.checker.HasPermission(ctx, userID, domain.PermFloodBypass); ok {
			return nil
		}
	}
	service.floodMu.Lock()
	defer service.floodMu.Unlock()
	state, ok := service.flood[connID]
	if !ok {
		state = &floodState{}
		service.flood[connID] = state
	}
	now := time.Now()
	if !state.mutedUntil.IsZero() && now.Before(state.mutedUntil) {
		return domain.ErrSenderMuted
	}
	state.mutedUntil = time.Time{}
	cooldown := time.Duration(service.config.FloodCooldownMs) * time.Millisecond
	if !state.lastMessage.IsZero() && now.Sub(state.lastMessage) < cooldown {
		state.violations++
		if state.violations >= service.config.FloodViolations {
			state.mutedUntil = now.Add(time.Duration(service.config.FloodMuteSeconds) * time.Second)
			state.violations = 0
		}
		return domain.ErrSenderMuted
	}
	state.lastMessage = now
	state.violations = 0
	return nil
}

// ClearFloodState removes flood tracking for one disconnected connection.
func (service *Service) ClearFloodState(connID string) {
	service.floodMu.Lock()
	defer service.floodMu.Unlock()
	delete(service.flood, connID)
}
