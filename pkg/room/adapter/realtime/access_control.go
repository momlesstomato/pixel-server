package realtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/room/domain"
	"github.com/momlesstomato/pixel-server/pkg/room/packet"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	notificationpacket "github.com/momlesstomato/pixel-server/pkg/session/packet/notification"
)

const (
	roomAccessAttemptLimit = 3
	roomAccessCooldown     = 30 * time.Second
	roomAccessPermission   = "pixels.room.access"
	roomMasterPermission   = "pixels.room.master"
	roomEnterPermission    = "room.enter"
	wrongPasswordErrorCode = -100002
	noRoomPresentMessage   = "no one is in the room right now"
)

// roomAccessAttempt stores password failure state for one user-room pair.
type roomAccessAttempt struct {
	// Failures stores the failed attempt count in the current window.
	Failures int
	// CooldownUntil stores when the password retry cooldown expires.
	CooldownUntil time.Time
}

// newAccessState creates one empty room access runtime state container.
func newAccessState() *accessState {
	return &accessState{
		pendingDoorbell:  make(map[string]doorbellEntry),
		passwordAttempts: make(map[string]roomAccessAttempt),
	}
}

// accessState stores mutable room entry state shared across connections.
type accessState struct {
	// mu protects all mutable access state maps.
	mu sync.Mutex
	// pendingDoorbell stores pending visitor approvals by username.
	pendingDoorbell map[string]doorbellEntry
	// passwordAttempts stores password failure tracking by user-room pair.
	passwordAttempts map[string]roomAccessAttempt
}

// roomAccessKey builds a stable access tracking key for one user-room pair.
func roomAccessKey(userID int, roomID int) string {
	return fmt.Sprintf("%d:%d", userID, roomID)
}

// canBypassRoomAccess reports whether one user may bypass room entry restrictions.
func (rt *Runtime) canBypassRoomAccess(ctx context.Context, userID int, room domain.Room) bool {
	if rt.permissions == nil {
		return false
	}
	scopes := []string{roomMasterPermission, roomAccessPermission, roomEnterPermission}
	switch room.State {
	case domain.AccessLocked:
		scopes = append(scopes, "room.enter.locked")
	case domain.AccessPassword:
		scopes = append(scopes, "room.enter.password")
	}
	for _, scope := range scopes {
		granted, err := rt.permissions.HasPermission(ctx, userID, scope)
		if err == nil && granted {
			return true
		}
	}
	return false
}

// canControlDoorbell reports whether one user may approve or deny doorbell access.
func (rt *Runtime) canControlDoorbell(ctx context.Context, room domain.Room, userID int) bool {
	if room.OwnerID == userID {
		return true
	}
	if rt.service.HasRights(ctx, room.ID, userID) {
		return true
	}
	return rt.canBypassRoomAccess(ctx, userID, room)
}

// currentCooldown returns the active password cooldown for one user-room pair.
func (rt *Runtime) currentCooldown(userID int, roomID int) (time.Duration, bool) {
	key := roomAccessKey(userID, roomID)
	rt.access.mu.Lock()
	defer rt.access.mu.Unlock()
	attempt, ok := rt.access.passwordAttempts[key]
	if !ok || time.Now().After(attempt.CooldownUntil) {
		if ok && !attempt.CooldownUntil.IsZero() {
			delete(rt.access.passwordAttempts, key)
		}
		return 0, false
	}
	return time.Until(attempt.CooldownUntil), true
}

// recordPasswordFailure increments password failures and returns the current cooldown, if any.
func (rt *Runtime) recordPasswordFailure(userID int, roomID int) (time.Duration, bool) {
	key := roomAccessKey(userID, roomID)
	rt.access.mu.Lock()
	defer rt.access.mu.Unlock()
	attempt := rt.access.passwordAttempts[key]
	now := time.Now()
	if now.Before(attempt.CooldownUntil) {
		rt.access.passwordAttempts[key] = attempt
		return time.Until(attempt.CooldownUntil), true
	}
	attempt.Failures++
	if attempt.Failures >= roomAccessAttemptLimit {
		attempt.Failures = 0
		attempt.CooldownUntil = now.Add(roomAccessCooldown)
		rt.access.passwordAttempts[key] = attempt
		return roomAccessCooldown, true
	}
	attempt.CooldownUntil = time.Time{}
	rt.access.passwordAttempts[key] = attempt
	return 0, false
}

// clearPasswordFailures resets password failure state for one user-room pair.
func (rt *Runtime) clearPasswordFailures(userID int, roomID int) {
	rt.access.mu.Lock()
	defer rt.access.mu.Unlock()
	delete(rt.access.passwordAttempts, roomAccessKey(userID, roomID))
}

// sendWrongPasswordFeedback sends the wrong-password denial packets.
func (rt *Runtime) sendWrongPasswordFeedback(connID string) error {
	if err := rt.sendPacket(connID, packet.CantConnectComposer{ErrorCode: 6}); err != nil {
		return err
	}
	return rt.sendPacket(connID, notificationpacket.GenericErrorPacket{ErrorCode: wrongPasswordErrorCode})
}

// sendPasswordCooldownFeedback sends cooldown feedback after repeated failures.
func (rt *Runtime) sendPasswordCooldownFeedback(connID string, retryAfter time.Duration) error {
	seconds := int32(retryAfter.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	if err := rt.sendWrongPasswordFeedback(connID); err != nil {
		return err
	}
	if err := rt.sendPacket(connID, packet.FloodControlComposer{Seconds: seconds}); err != nil {
		return err
	}
	if err := rt.sendPacket(connID, sessionnavigation.DesktopViewResponsePacket{}); err != nil {
		return err
	}
	return rt.sendPacket(connID, notificationpacket.GenericAlertPacket{
		Message: fmt.Sprintf("Too many wrong password attempts. Please wait %d seconds before trying again.", seconds),
	})
}

// sendNoRoomPresentFeedback returns the user to hotel view with an explicit empty-room message.
func (rt *Runtime) sendNoRoomPresentFeedback(connID string) error {
	if err := rt.sendPacket(connID, sessionnavigation.DesktopViewResponsePacket{}); err != nil {
		return err
	}
	return rt.sendPacket(connID, notificationpacket.GenericAlertPacket{Message: noRoomPresentMessage})
}

// publishPacketToUser pushes one packet to a user notification channel.
func (rt *Runtime) publishPacketToUser(ctx context.Context, userID int, pkt interface {
	PacketID() uint16
	Encode() ([]byte, error)
}) {
	if rt.broadcaster == nil || userID <= 0 {
		return
	}
	body, err := pkt.Encode()
	if err != nil {
		return
	}
	_ = rt.broadcaster.Publish(ctx, sessionnotification.UserChannel(userID), codec.EncodeFrame(pkt.PacketID(), body))
}

// cleanupDoorbellForConn removes any pending doorbell entry for one connection.
func (rt *Runtime) cleanupDoorbellForConn(connID string) {
	rt.access.mu.Lock()
	defer rt.access.mu.Unlock()
	for username, entry := range rt.access.pendingDoorbell {
		if entry.ConnID == connID {
			delete(rt.access.pendingDoorbell, username)
		}
	}
}
