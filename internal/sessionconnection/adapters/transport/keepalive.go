package transport

import (
	"context"
	"time"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
)

// runPingLoop publishes periodic client.ping frames to authenticated sessions.
func (s *Subscriber) runPingLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.PingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.broadcastPing(ctx)
		}
	}
}

// runTimeoutLoop disconnects stale sessions that miss pong deadlines.
func (s *Subscriber) runTimeoutLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.disconnectExpiredSessions(ctx, now)
		}
	}
}

// broadcastPing publishes one keepalive ping to each authenticated session.
func (s *Subscriber) broadcastPing(ctx context.Context) {
	for _, sessionID := range s.service.ActiveAuthenticatedSessions() {
		if err := s.publishOutput(ctx, sessionID, encodePing()); err != nil {
			s.logger.Warn("failed to publish ping", zap.String("session_id", sessionID), zap.Error(err))
		}
	}
}

// disconnectExpiredSessions disconnects sessions evicted by pong timeout policy.
func (s *Subscriber) disconnectExpiredSessions(ctx context.Context, now time.Time) {
	for _, sessionID := range s.service.ExpirePongTimeoutSessions(s.cfg.PongTimeout, now) {
		if err := s.disconnectSession(ctx, sessionID, sessionmessaging.DisconnectReasonIdleTimeout); err != nil {
			s.logger.Warn("failed to disconnect stale session", zap.String("session_id", sessionID), zap.Error(err))
		}
	}
}

// encodeAvailabilityStatus encodes one availability.status frame.
func encodeAvailabilityStatus(cfg Config) []byte {
	writer := codec.NewWriter(12)
	writer.WriteBool(cfg.AvailabilityOpen)
	writer.WriteBool(cfg.AvailabilityOnShutdown)
	writer.WriteBool(cfg.AvailabilityAuthentic)
	return codec.EncodeFrame(2033, writer.Bytes())
}

// encodeLatencyResponse encodes one client.latency_response frame.
func encodeLatencyResponse(requestID int32) []byte {
	writer := codec.NewWriter(8)
	writer.WriteInt32(requestID)
	return codec.EncodeFrame(10, writer.Bytes())
}

// encodeDisconnectReason encodes one disconnect.reason frame.
func encodeDisconnectReason(reason int32) []byte {
	writer := codec.NewWriter(8)
	writer.WriteInt32(reason)
	return codec.EncodeFrame(4000, writer.Bytes())
}

// encodeDesktopViewAck encodes one session.desktop_view ack frame.
func encodeDesktopViewAck() []byte {
	return codec.EncodeFrame(122, nil)
}

// encodePing encodes one client.ping frame.
func encodePing() []byte {
	return codec.EncodeFrame(3928, nil)
}
