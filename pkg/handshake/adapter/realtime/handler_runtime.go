package realtime

import (
	"context"

	"github.com/gofiber/contrib/websocket"
	packetcrypto "github.com/momlesstomato/pixel-server/pkg/handshake/packet/crypto"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
	packettelemetry "github.com/momlesstomato/pixel-server/pkg/handshake/packet/telemetry"
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
	authSignal, pongSignal, heartbeatStop := make(chan struct{}), make(chan struct{}, 1), func() {}
	defer handler.disposeConnection(cancel, disposable, useCases.disconnect, heartbeatStop, connID, connection)
	go func() {
		if useCases.timeout.Wait(ctx, connID, authSignal) != nil {
			cancel()
		}
	}()
	go func() {
		select {
		case signal, open := <-signals:
			if open {
				_ = transport.closeLocal(signal.Code, signal.Reason)
			}
			cancel()
		case <-ctx.Done():
		}
	}()
	handler.readLoop(ctx, connection, transport, useCases, connID, authSignal, pongSignal, &heartbeatStop, cancel)
}

// readLoop reads websocket packets and applies handshake session workflows.
func (handler *Handler) readLoop(ctx context.Context, connection *websocket.Conn, transport *Transport, useCases *runtimeUseCases, connID string, authSignal chan struct{}, pongSignal chan struct{}, heartbeatStop *func(), cancel context.CancelFunc) {
	authenticated, machineID := false, ""
	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		frames, err := transport.DecodeFrames(payload)
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
				if !handler.handleSSO(ctx, connID, frame.Body, machineID, useCases.authenticate) {
					return
				}
				if !authenticated {
					authenticated = true
					close(authSignal)
					*heartbeatStop = handler.startHeartbeat(ctx, connID, useCases.heartbeat, pongSignal, cancel)
				}
			case packetsession.ClientDisconnectPacketID:
				if handler.handleDisconnect(connID, frame.Body, useCases.disconnect) {
					return
				}
			case packetsession.ClientPongPacketID:
				if authenticated {
					packet := packetsession.ClientPongPacket{}
					if packet.Decode(frame.Body) == nil {
						select {
						case pongSignal <- struct{}{}:
						default:
						}
					}
				}
			case packettelemetry.ClientLatencyTestPacketID:
				if authenticated {
					packet := packettelemetry.ClientLatencyTestPacket{}
					if packet.Decode(frame.Body) == nil {
						_ = useCases.latency.Respond(connID, packet.RequestID)
					}
				}
			}
		}
	}
}
