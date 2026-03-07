package ws

import (
	"context"
	"encoding/binary"
	"errors"

	"go.uber.org/zap"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	"pixelsv/pkg/core/session"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/protocol"
)

// handleBinary decodes one websocket payload and publishes packet events.
func (g *Gateway) handleBinary(ctx context.Context, sessionID string, raw []byte) error {
	frames, err := codec.SplitFrames(raw)
	if err != nil {
		return err
	}
	for _, frame := range frames {
		packet, err := protocol.DecodeC2S(frame.Header, frame.Payload)
		if err != nil {
			if errors.Is(err, protocol.ErrUnknownHeader) {
				g.logger.Debug(
					"websocket unknown packet header",
					zap.String("session_id", sessionID),
					zap.Uint16("header", frame.Header),
					zap.Int("payload_bytes", len(frame.Payload)),
				)
				continue
			}
			return err
		}
		topic := transport.PacketC2STopic(packet.Realm(), sessionID)
		g.logger.Debug(
			"websocket packet ingress",
			zap.String("session_id", sessionID),
			zap.Uint16("header", frame.Header),
			zap.String("packet", packet.PacketName()),
			zap.String("topic", topic),
			zap.Int("payload_bytes", len(frame.Payload)),
		)
		if err := g.bus.Publish(ctx, topic, frame.Body); err != nil {
			return err
		}
	}
	return nil
}

// handleSessionOutput writes transport output payloads to websocket sessions.
func (g *Gateway) handleSessionOutput(ctx context.Context, message transport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID, ok := sessionmessaging.ParseOutputTopic(message.Topic)
	if !ok {
		return nil
	}
	if err := g.sessions.Send(sessionID, message.Payload); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	g.logger.Debug(
		"websocket packet egress",
		zap.String("session_id", sessionID),
		zap.String("topic", message.Topic),
		zap.Int("payload_bytes", len(message.Payload)),
	)
	return nil
}

// handleSessionDisconnect closes websocket sessions requested by runtime control topics.
func (g *Gateway) handleSessionDisconnect(ctx context.Context, message transport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID, ok := sessionmessaging.ParseDisconnectTopic(message.Topic)
	if !ok {
		return nil
	}
	reason := int32(0)
	if len(message.Payload) >= 4 {
		reason = int32(binary.BigEndian.Uint32(message.Payload[:4]))
	}
	if err := g.sessions.Remove(sessionID); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	g.logger.Info("websocket session closed by runtime", zap.String("session_id", sessionID), zap.Int32("reason", reason))
	return nil
}
