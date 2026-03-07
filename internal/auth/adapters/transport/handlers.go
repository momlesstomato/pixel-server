package transport

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"pixelsv/internal/auth/app"
	"pixelsv/internal/auth/domain"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/protocol"
)

// dispatchPacket routes one decoded handshake packet.
func (s *Subscriber) dispatchPacket(ctx context.Context, sessionID string, packet protocol.Packet) error {
	switch value := packet.(type) {
	case *protocol.HandshakeReleaseVersionPacket:
		return s.handleReleaseVersion(ctx, sessionID, value)
	case *protocol.HandshakeClientVariablesPacket:
		return s.service.RecordClientVariables(sessionID, value)
	case *protocol.HandshakeInitDiffiePacket:
		return s.handleInitDiffie(ctx, sessionID)
	case *protocol.HandshakeCompleteDiffiePacket:
		return s.handleCompleteDiffie(ctx, sessionID, value)
	case *protocol.SecurityMachineIdPacket:
		return s.handleMachineID(ctx, sessionID, value)
	case *protocol.HandshakeClientLatencyMeasurePacket:
		return s.service.MarkLatencyMeasure(sessionID)
	case *protocol.HandshakeClientPolicyPacket:
		return s.service.MarkClientPolicy(sessionID)
	case *protocol.SecuritySsoTicketPacket:
		return s.handleSSOTicket(ctx, sessionID, value.Ticket)
	default:
		return nil
	}
}

// handleReleaseVersion validates release metadata and disconnects unsupported clients.
func (s *Subscriber) handleReleaseVersion(ctx context.Context, sessionID string, packet *protocol.HandshakeReleaseVersionPacket) error {
	if err := s.service.RecordReleaseVersion(sessionID, packet); err != nil {
		return s.rejectSession(ctx, sessionID, 3, err)
	}
	return nil
}

// handleInitDiffie publishes init_diffie response values.
func (s *Subscriber) handleInitDiffie(ctx context.Context, sessionID string) error {
	response, err := s.service.InitDiffie(sessionID)
	if err != nil {
		return s.rejectSession(ctx, sessionID, 4, err)
	}
	return s.publishOutput(ctx, sessionID, encodeInitDiffieFrame(response))
}

// handleCompleteDiffie publishes complete_diffie response values.
func (s *Subscriber) handleCompleteDiffie(ctx context.Context, sessionID string, packet *protocol.HandshakeCompleteDiffiePacket) error {
	response, err := s.service.CompleteDiffie(sessionID, packet.EncryptedPublicKey)
	if err != nil {
		return s.rejectSession(ctx, sessionID, 4, err)
	}
	return s.publishOutput(ctx, sessionID, encodeCompleteDiffieFrame(response))
}

// handleMachineID normalizes and optionally echoes machine id values.
func (s *Subscriber) handleMachineID(ctx context.Context, sessionID string, packet *protocol.SecurityMachineIdPacket) error {
	normalized, changed, err := s.service.UpdateMachineID(sessionID, packet.MachineId, packet.Fingerprint, packet.Capabilities)
	if err != nil {
		return s.rejectSession(ctx, sessionID, 5, err)
	}
	if !changed {
		return nil
	}
	return s.publishOutput(ctx, sessionID, encodeMachineIDFrame(normalized))
}

// handleSSOTicket validates one SSO ticket and emits auth transport events.
func (s *Subscriber) handleSSOTicket(ctx context.Context, sessionID string, ticket string) error {
	userID, err := s.service.ValidateTicket(sessionID, ticket)
	if err != nil {
		if errors.Is(err, domain.ErrTicketNotFound) || errors.Is(err, domain.ErrInvalidTicket) {
			return s.rejectSession(ctx, sessionID, 1, err)
		}
		if errors.Is(err, app.ErrReleaseVersionRequired) || errors.Is(err, app.ErrDiffieRequired) {
			return s.rejectSession(ctx, sessionID, 2, err)
		}
		return err
	}
	authenticated := sessionmessaging.EncodeAuthenticatedEvent(sessionID, userID)
	if err := s.bus.Publish(ctx, sessionmessaging.TopicAuthenticated, authenticated); err != nil {
		return err
	}
	return s.publishOutput(ctx, sessionID, encodeAuthSuccessFrames())
}

// handleExpiredSessions disconnects handshake sessions that timed out.
func (s *Subscriber) handleExpiredSessions(ctx context.Context, now time.Time) {
	for _, sessionID := range s.service.ExpireUnauthenticatedSessions(now) {
		err := s.publishOutput(ctx, sessionID, encodeConnectionErrorFrame(4))
		if err == nil {
			err = s.publishDisconnect(ctx, sessionID, 4)
		}
		if err != nil {
			s.logger.Warn("failed to disconnect expired handshake session", zap.String("session_id", sessionID), zap.Error(err))
		}
	}
}
