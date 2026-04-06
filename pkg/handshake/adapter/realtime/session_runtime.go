package realtime

import (
	"context"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/authentication"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	packeterror "github.com/momlesstomato/pixel-server/pkg/session/packet/error"
	packetsnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
)

// protocolErrorMeter defines connection.error rate-limit behavior.
type protocolErrorMeter struct {
	// startedAt stores beginning timestamp for active accounting window.
	startedAt time.Time
	// count stores emitted connection.error count in current window.
	count int
}

// newProtocolErrorMeter creates one protocol error meter instance.
func newProtocolErrorMeter() protocolErrorMeter {
	return protocolErrorMeter{startedAt: time.Now().UTC(), count: 0}
}

// register increments accounting counters and reports whether limit is exceeded.
func (meter *protocolErrorMeter) register(now time.Time) bool {
	if now.Sub(meter.startedAt) >= time.Minute {
		meter.startedAt, meter.count = now, 0
	}
	meter.count++
	return meter.count > protocolErrorLimitPerMinute
}

// handleProtocolError sends connection.error and closes on error-flood threshold.
func (handler *Handler) handleProtocolError(connID string, transport *Transport, messageID uint16, errorCode int32, meter *protocolErrorMeter) bool {
	handler.sendConnectionError(connID, transport, messageID, errorCode)
	if meter.register(time.Now().UTC()) {
		_ = transport.Close(connID, websocket.ClosePolicyViolation, "protocol error flood")
		return true
	}
	return false
}

// subscribeChannel subscribes one broadcast channel when broadcaster is configured.
func (handler *Handler) subscribeChannel(ctx context.Context, channel string) (<-chan []byte, coreconnection.Disposable, error) {
	if handler.broadcaster == nil {
		return nil, nil, nil
	}
	return handler.broadcaster.Subscribe(ctx, channel)
}

// consumeBroadcast delivers distributed broadcast payloads to active websocket transport.
func (handler *Handler) consumeBroadcast(ctx context.Context, stream <-chan []byte, connID string, transport *Transport, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case payload, open := <-stream:
			if !open {
				return
			}
			if !handler.deliverBroadcastPayload(connID, transport, payload) {
				cancel()
				return
			}
		}
	}
}

// deliverBroadcastPayload forwards one broadcast payload and handles disconnect reasons.
func (handler *Handler) deliverBroadcastPayload(connID string, transport *Transport, payload []byte) bool {
	frames, err := codec.DecodeFrames(payload)
	if err != nil {
		return true
	}
	for _, frame := range frames {
		if frame.PacketID != packetauth.DisconnectReasonPacketID {
			_ = transport.writeFrame(frame.PacketID, frame.Body)
			continue
		}
		reason := packetauth.DisconnectReasonPacket{}
		if reason.Decode(frame.Body) != nil {
			continue
		}
		_ = transport.writeFrame(frame.PacketID, frame.Body)
		code, closeReason := disconnectClose(reason.Reason)
		_ = transport.Close(connID, code, closeReason)
		return false
	}
	return true
}

// disconnectClose maps protocol disconnect reasons to websocket close metadata.
func disconnectClose(reason int32) (int, string) {
	if reason == packetauth.DisconnectReasonHotelClosed {
		return websocket.ClosePolicyViolation, "hotel closed"
	}
	if reason == packetauth.DisconnectReasonJustBanned || reason == packetauth.DisconnectReasonStillBanned {
		return websocket.ClosePolicyViolation, "banned"
	}
	return websocket.CloseNormalClosure, "disconnect"
}

// handleDesktopView handles one desktop view request packet using navigation use case.
func (handler *Handler) handleDesktopView(ctx context.Context, connID string, body []byte, useCase interface {
	Run(context.Context, string, int) error
}) bool {
	packet := packetsnavigation.DesktopViewRequestPacket{}
	if packet.Decode(body) != nil {
		return false
	}
	session, found := handler.sessions.FindByConnID(connID)
	return found && session.UserID > 0 && useCase.Run(ctx, connID, session.UserID) == nil
}

// sendConnectionError sends one connection.error packet to active connection.
func (handler *Handler) sendConnectionError(connID string, transport *Transport, messageID uint16, errorCode int32) {
	packet := packeterror.ConnectionErrorPacket{MessageID: int32(messageID), ErrorCode: errorCode, Timestamp: time.Now().UTC().Format(time.RFC3339Nano)}
	if body, err := packet.Encode(); err == nil {
		_ = transport.Send(connID, packet.PacketID(), body)
	}
}

// handleAuthPacket authenticates one connection and wires post-auth behavior.
func (handler *Handler) handleAuthPacket(ctx context.Context, connID string, body []byte, machineID string, transport *Transport, useCases *runtimeUseCases, authSignal chan struct{}, pongSignal chan struct{}, heartbeatStop *func(), disposables *[]coreconnection.Disposable, cancel context.CancelFunc, userSubscribed *bool, userRuntime UserRuntime) bool {
	userID, ok := handler.handleSSO(ctx, connID, body, machineID, useCases.authenticate)
	if !ok {
		return false
	}
	if useCases.postauth != nil && useCases.postauth.Run(ctx, connID, userID) != nil {
		return false
	}
	if hook, ok := userRuntime.(PostAuthHook); ok {
		hook.OnPostAuth(ctx, connID, userID)
	}
	if !*userSubscribed {
		userMessages, userDisposable, err := handler.subscribeChannel(ctx, sessionnotification.UserChannel(userID))
		if err != nil {
			return false
		}
		if userDisposable != nil {
			*disposables = append(*disposables, userDisposable)
			go handler.consumeBroadcast(ctx, userMessages, connID, transport, cancel)
			*userSubscribed = true
		}
	}
	close(authSignal)
	*heartbeatStop = handler.startHeartbeat(ctx, connID, useCases.heartbeat, pongSignal, cancel)
	return true
}
