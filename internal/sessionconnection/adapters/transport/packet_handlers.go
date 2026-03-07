package transport

import (
	"context"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/protocol"
)

// handlePacket routes one session-connection packet ingress event.
func (s *Subscriber) handlePacket(ctx context.Context, message coretransport.Message) error {
	sessionID, ok := sessionmessaging.ParsePacketIngressTopic(message.Topic)
	if !ok {
		return nil
	}
	packet, err := decodePacket(message.Payload)
	if err != nil || packet == nil {
		return err
	}
	s.logger.Debug(
		"session-connection packet received",
		zap.String("session_id", sessionID),
		zap.Uint16("header", packet.HeaderID()),
		zap.String("packet", packet.PacketName()),
	)
	if err := ignoreMissing(s.service.RecordPacket(sessionID, packet.HeaderID(), packet.PacketName())); err != nil {
		return err
	}
	switch value := packet.(type) {
	case *protocol.ClientLatencyTestPacket:
		return s.handleLatencyTest(ctx, sessionID, value.RequestId)
	case *protocol.ClientPongPacket:
		return ignoreMissing(s.service.MarkPong(sessionID))
	case *protocol.ClientDisconnectPacket:
		return s.disconnectSession(ctx, sessionID, sessionmessaging.DisconnectReasonGeneric)
	case *protocol.SessionDesktopViewPacket:
		return s.handleDesktopView(ctx, sessionID)
	case *protocol.SessionPeerUsersClassificationPacket, *protocol.SessionClientToolbarTogglePacket, *protocol.SessionRenderRoomPacket:
		return nil
	case *protocol.SessionTrackingPerformanceLogPacket, *protocol.SessionEventTrackerPacket, *protocol.SessionTrackingLagWarningReportPacket:
		return s.handleTelemetry(sessionID, packet.HeaderID())
	default:
		return nil
	}
}

// handleLatencyTest records one latency request and publishes echo response.
func (s *Subscriber) handleLatencyTest(ctx context.Context, sessionID string, requestID int32) error {
	if err := ignoreMissing(s.service.MarkLatencyTest(sessionID, requestID)); err != nil {
		return err
	}
	return s.publishOutput(ctx, sessionID, encodeLatencyResponse(requestID))
}

// handleDesktopView records desktop view signal and publishes ack response.
func (s *Subscriber) handleDesktopView(ctx context.Context, sessionID string) error {
	if err := ignoreMissing(s.service.MarkDesktopView(sessionID)); err != nil {
		return err
	}
	return s.publishOutput(ctx, sessionID, encodeDesktopViewAck())
}

// handleTelemetry applies throttling and logs accepted telemetry packets.
func (s *Subscriber) handleTelemetry(sessionID string, header uint16) error {
	allowed, err := s.service.AllowTelemetry(sessionID, header)
	if err != nil {
		return ignoreMissing(err)
	}
	if allowed {
		s.logger.Debug("session telemetry packet processed", zap.String("session_id", sessionID), zap.Uint16("header", header))
	}
	return nil
}
