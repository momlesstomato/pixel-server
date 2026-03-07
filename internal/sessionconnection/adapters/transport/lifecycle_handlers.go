package transport

import (
	"context"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	coretransport "pixelsv/pkg/core/transport"
)

// handleConnected initializes session state for newly connected websocket sessions.
func (s *Subscriber) handleConnected(ctx context.Context, message coretransport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID := string(message.Payload)
	if err := s.service.SessionConnected(sessionID); err != nil {
		return err
	}
	return nil
}

// handleDisconnected removes session state after gateway disconnect lifecycle events.
func (s *Subscriber) handleDisconnected(ctx context.Context, message coretransport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID := string(message.Payload)
	s.service.SessionDisconnected(sessionID)
	s.logger.Info("session-connection cleanup handled", zap.String("session_id", sessionID))
	return nil
}

// handleAuthenticated initializes post-auth session flow and availability output.
func (s *Subscriber) handleAuthenticated(ctx context.Context, message coretransport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := sessionmessaging.DecodeAuthenticatedEvent(message.Payload)
	if err != nil {
		return err
	}
	previousSessionID, err := s.service.SessionAuthenticated(payload.SessionID, payload.UserID)
	if err != nil {
		return err
	}
	if previousSessionID != "" && previousSessionID != payload.SessionID {
		if err := s.disconnectSession(ctx, previousSessionID, sessionmessaging.DisconnectReasonConcurrentLogin); err != nil {
			return err
		}
		s.logger.Info("session disconnected due to concurrent login", zap.String("session_id", previousSessionID), zap.Int32("reason", sessionmessaging.DisconnectReasonConcurrentLogin))
	}
	if err := s.publishOutput(ctx, payload.SessionID, encodeAvailabilityStatus(s.cfg)); err != nil {
		return err
	}
	return nil
}
