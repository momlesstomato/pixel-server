package transport

import (
	"context"

	"go.uber.org/zap"
	"pixelsv/internal/auth/app"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
)

// publishOutput sends one binary payload to one websocket session output topic.
func (s *Subscriber) publishOutput(ctx context.Context, sessionID string, payload []byte) error {
	s.logger.Debug("auth output publish", zap.String("session_id", sessionID), zap.Int("payload_bytes", len(payload)))
	return s.bus.Publish(ctx, sessionmessaging.OutputTopic(sessionID), payload)
}

// publishDisconnect asks gateway to close one websocket session by reason code.
func (s *Subscriber) publishDisconnect(ctx context.Context, sessionID string, reason int32) error {
	writer := codec.NewWriter(8)
	writer.WriteInt32(reason)
	s.logger.Debug("auth disconnect publish", zap.String("session_id", sessionID), zap.Int32("reason", reason))
	return s.bus.Publish(ctx, sessionmessaging.DisconnectTopic(sessionID), writer.Bytes())
}

// rejectSession publishes connection error and disconnect control for one session.
func (s *Subscriber) rejectSession(ctx context.Context, sessionID string, reason int32, cause error) error {
	s.logger.Warn("handshake session rejected", zap.String("session_id", sessionID), zap.Int32("reason", reason), zap.Error(cause))
	s.service.RemoveSession(sessionID)
	if err := s.publishOutput(ctx, sessionID, encodeConnectionErrorFrame(reason)); err != nil {
		return err
	}
	return s.publishDisconnect(ctx, sessionID, reason)
}

// encodeInitDiffieFrame encodes one init_diffie server frame.
func encodeInitDiffieFrame(response app.InitDiffieResponse) []byte {
	writer := codec.NewWriter(320)
	writer.WriteString(response.SignedPrime)
	writer.WriteString(response.SignedGenerator)
	return codec.EncodeFrame(1347, writer.Bytes())
}

// encodeCompleteDiffieFrame encodes one complete_diffie server frame.
func encodeCompleteDiffieFrame(response app.CompleteDiffieResponse) []byte {
	writer := codec.NewWriter(192)
	writer.WriteString(response.PublicKey)
	writer.WriteBool(response.ServerEncryption)
	return codec.EncodeFrame(3885, writer.Bytes())
}

// encodeMachineIDFrame encodes one machine-id server frame.
func encodeMachineIDFrame(machineID string) []byte {
	writer := codec.NewWriter(96)
	writer.WriteString(machineID)
	return codec.EncodeFrame(1488, writer.Bytes())
}

// encodeConnectionErrorFrame encodes one connection error server frame.
func encodeConnectionErrorFrame(reason int32) []byte {
	writer := codec.NewWriter(8)
	writer.WriteInt32(reason)
	return codec.EncodeFrame(1004, writer.Bytes())
}

// encodeAuthSuccessFrames encodes auth-ok and identity-account frames.
func encodeAuthSuccessFrames() []byte {
	identity := codec.NewWriter(8)
	identity.WriteInt32(0)
	authOK := codec.EncodeFrame(2491, nil)
	accounts := codec.EncodeFrame(3523, identity.Bytes())
	payload := make([]byte, 0, len(authOK)+len(accounts))
	payload = append(payload, authOK...)
	payload = append(payload, accounts...)
	return payload
}
