package ws

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/core/session"
)

var shutdownDisconnectFlushDelay = 200 * time.Millisecond

// Shutdown disconnects every active websocket session and publishes lifecycle cleanup events.
func (g *Gateway) Shutdown(ctx context.Context, reason int32) {
	sessionIDs := g.sessions.IDs()
	for _, sessionID := range sessionIDs {
		if err := g.sessions.Send(sessionID, disconnectReasonFrame(reason)); err != nil && !errors.Is(err, session.ErrSessionNotFound) {
			g.logger.Debug("failed to send shutdown disconnect.reason", zap.String("session_id", sessionID), zap.Error(err))
		}
		if err := g.bus.Publish(ctx, sessionmessaging.TopicDisconnected, []byte(sessionID)); err != nil {
			g.logger.Debug("failed to publish shutdown session.disconnected", zap.String("session_id", sessionID), zap.Error(err))
		}
	}
	if shutdownDisconnectFlushDelay > 0 {
		time.Sleep(shutdownDisconnectFlushDelay)
	}
	for _, sessionID := range sessionIDs {
		if err := g.sessions.Remove(sessionID); err != nil && !errors.Is(err, session.ErrSessionNotFound) {
			g.logger.Debug("failed to remove shutdown session", zap.String("session_id", sessionID), zap.Error(err))
		}
		g.logger.Info("websocket session disconnected by shutdown", zap.String("session_id", sessionID), zap.Int32("reason", reason))
	}
}
