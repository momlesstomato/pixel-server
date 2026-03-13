package realtime

import (
	"context"

	"github.com/gofiber/contrib/websocket"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
	packettelemetry "github.com/momlesstomato/pixel-server/pkg/handshake/packet/telemetry"
	sessionnotification "github.com/momlesstomato/pixel-server/pkg/session/application/notification"
	packetsnavigation "github.com/momlesstomato/pixel-server/pkg/session/packet/navigation"
	"go.uber.org/zap"
)

// Handle executes websocket handshake packet workflow for one connection.
func (handler *Handler) Handle(connection *websocket.Conn) {
	connID, err := handler.connID()
	if err != nil {
		handler.abortConnection(connection)
		return
	}
	transport, err := NewTransport(connID, connection, handler.bus, handler.logger)
	if err != nil {
		handler.abortConnection(connection)
		return
	}
	useCases, err := handler.newRuntimeUseCases(transport)
	if err != nil {
		handler.abortConnection(connection)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	signals, disposable, err := handler.bus.Subscribe(ctx, connID)
	if err != nil {
		cancel()
		handler.abortConnection(connection)
		return
	}
	disposables := []coreconnection.Disposable{disposable}
	allMessages, allDisposable, err := handler.subscribeChannel(ctx, sessionnotification.AllChannel())
	if err != nil {
		cancel()
		handler.abortConnection(connection)
		return
	}
	if allDisposable != nil {
		disposables = append(disposables, allDisposable)
		go handler.consumeBroadcast(ctx, allMessages, connID, transport, cancel)
	}
	authSignal, pongSignal, heartbeatStop := make(chan struct{}), make(chan struct{}, 1), func() {}
	defer handler.disposeConnection(cancel, disposables, useCases.disconnect, heartbeatStop, connID, connection)
	handler.startRuntimeWatchers(ctx, useCases, connID, authSignal, signals, transport, cancel)
	handler.readLoop(ctx, connection, transport, useCases, connID, authSignal, pongSignal, &heartbeatStop, &disposables, cancel)
}

// readLoop reads websocket packets and applies handshake session workflows.
func (handler *Handler) readLoop(ctx context.Context, connection *websocket.Conn, transport *Transport, useCases *runtimeUseCases, connID string, authSignal chan struct{}, pongSignal chan struct{}, heartbeatStop *func(), disposables *[]coreconnection.Disposable, cancel context.CancelFunc) {
	authenticated, machineID, userSubscribed := false, "", false
	errorMeter := newProtocolErrorMeter()
	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		frames, err := transport.DecodeFrames(payload)
		if err != nil && handler.handleProtocolError(connID, transport, 0, protocolErrorMalformedPacket, &errorMeter) {
			return
		}
		if err != nil {
			continue
		}
		for _, frame := range frames {
			handler.logger.Debug("websocket packet received", zap.String("conn_id", connID), zap.Uint16("packet_id", frame.PacketID), zap.Int("size", len(frame.Body)))
			switch frame.PacketID {
			case packetauth.ClientMachineIDPacketID:
				handler.handleMachineID(connID, frame.Body, transport, &machineID)
			case packetcrypto.ClientInitDiffiePacketID:
				handler.handleInitDiffie(connID, frame.Body, transport, useCases)
			case packetcrypto.ClientCompleteDiffiePacketID:
				handler.handleCompleteDiffie(connID, frame.Body, transport, useCases)
			case packetauth.SSOTicketPacketID:
				if authenticated {
					if handler.handleProtocolError(connID, transport, frame.PacketID, protocolErrorWrongState, &errorMeter) {
						return
					}
					continue
				}
				if !handler.handleAuthPacket(ctx, connID, frame.Body, machineID, transport, useCases, authSignal, pongSignal, heartbeatStop, disposables, cancel, &userSubscribed) {
					return
				}
				authenticated = true
			case packetsession.ClientDisconnectPacketID:
				if handler.handleDisconnect(connID, frame.Body, useCases.disconnect) {
					return
				}
			case packetsession.ClientPongPacketID:
				packet := packetsession.ClientPongPacket{}
				if !authenticated || packet.Decode(frame.Body) != nil {
					errorCode := protocolErrorWrongState
					if authenticated {
						errorCode = protocolErrorMalformedPacket
					}
					if handler.handleProtocolError(connID, transport, frame.PacketID, errorCode, &errorMeter) {
						return
					}
					continue
				}
				select {
				case pongSignal <- struct{}{}:
				default:
				}
			case packettelemetry.ClientLatencyTestPacketID:
				packet := packettelemetry.ClientLatencyTestPacket{}
				if !authenticated || packet.Decode(frame.Body) != nil {
					errorCode := protocolErrorWrongState
					if authenticated {
						errorCode = protocolErrorMalformedPacket
					}
					if handler.handleProtocolError(connID, transport, frame.PacketID, errorCode, &errorMeter) {
						return
					}
					continue
				}
				_ = useCases.latency.Respond(connID, packet.RequestID)
			case packetsnavigation.DesktopViewRequestPacketID:
				if !authenticated {
					if handler.handleProtocolError(connID, transport, frame.PacketID, protocolErrorWrongState, &errorMeter) {
						return
					}
					continue
				}
				if useCases.desktop != nil && !handler.handleDesktopView(ctx, connID, frame.Body, useCases.desktop) &&
					handler.handleProtocolError(connID, transport, frame.PacketID, protocolErrorMalformedPacket, &errorMeter) {
					return
				}
			default:
				if handler.handleProtocolError(connID, transport, frame.PacketID, protocolErrorUnknownPacket, &errorMeter) {
					return
				}
			}
		}
	}
}
