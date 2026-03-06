package ws

import (
	"context"
	"errors"

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
			return err
		}
		topic := transport.PacketC2STopic(packet.Realm(), sessionID)
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
	sessionID, ok := transport.ParseSessionOutputTopic(message.Topic)
	if !ok {
		return nil
	}
	if err := g.sessions.Send(sessionID, message.Payload); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return nil
}
