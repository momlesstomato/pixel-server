package transport

import (
	"context"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
)

// publishOutput publishes one binary payload to session output topic.
func (s *Subscriber) publishOutput(ctx context.Context, sessionID string, payload []byte) error {
	s.logger.Debug("session output publish", zap.String("session_id", sessionID), zap.Int("payload_bytes", len(payload)))
	return s.bus.Publish(ctx, sessionmessaging.OutputTopic(sessionID), payload)
}

// publishDisconnect publishes one disconnect control signal with reason code.
func (s *Subscriber) publishDisconnect(ctx context.Context, sessionID string, reason int32) error {
	writer := codec.NewWriter(8)
	writer.WriteInt32(reason)
	s.logger.Debug("session disconnect publish", zap.String("session_id", sessionID), zap.Int32("reason", reason))
	return s.bus.Publish(ctx, sessionmessaging.DisconnectTopic(sessionID), writer.Bytes())
}

// disconnectSession publishes disconnect.reason output and disconnect control.
func (s *Subscriber) disconnectSession(ctx context.Context, sessionID string, reason int32) error {
	if err := s.publishOutput(ctx, sessionID, encodeDisconnectReason(reason)); err != nil {
		return err
	}
	return s.publishDisconnect(ctx, sessionID, reason)
}
