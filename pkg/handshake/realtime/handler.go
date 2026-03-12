package realtime

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/handshake/authflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	"go.uber.org/zap"
)

// Handler defines websocket handshake runtime behavior.
type Handler struct {
	// validator validates SSO tickets during authentication flow.
	validator authflow.TicketValidator
	// sessions stores connection lifecycle state in the session registry.
	sessions authflow.SessionRegistry
	// policy validates and regenerates machine identifiers.
	policy *packetauth.MachineIDPolicy
	// bus publishes and receives distributed close instructions.
	bus CloseSignalBus
	// logger stores runtime structured log behavior.
	logger *zap.Logger
	// authTimeout stores authentication timeout duration.
	authTimeout time.Duration
	// connID creates stable connection identifiers.
	connID func() (string, error)
}

// NewHandler creates websocket handshake runtime behavior.
func NewHandler(validator authflow.TicketValidator, sessions authflow.SessionRegistry, policy *packetauth.MachineIDPolicy, bus CloseSignalBus, logger *zap.Logger, authTimeout time.Duration) (*Handler, error) {
	if validator == nil {
		return nil, fmt.Errorf("ticket validator is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if bus == nil {
		return nil, fmt.Errorf("close signal bus is required")
	}
	if policy == nil {
		policy = packetauth.NewMachineIDPolicy(nil)
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Handler{validator: validator, sessions: sessions, policy: policy, bus: bus, logger: logger, authTimeout: authTimeout, connID: func() (string, error) { return GenerateConnectionID(rand.Reader) }}, nil
}

// Handle executes websocket handshake packet workflow for one connection.
func (handler *Handler) Handle(connection *websocket.Conn) {
	connID, err := handler.connID()
	if err != nil {
		_ = connection.Close()
		return
	}
	transport, err := NewTransport(connID, connection, handler.bus, handler.logger)
	if err != nil {
		_ = connection.Close()
		return
	}
	useCase, err := authflow.NewAuthenticateUseCase(handler.validator, handler.sessions, transport)
	if err != nil {
		_ = connection.Close()
		return
	}
	timeoutUseCase, err := authflow.NewTimeoutUseCase(transport, handler.authTimeout)
	if err != nil {
		_ = connection.Close()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signals, disposable, err := handler.bus.Subscribe(ctx, connID)
	if err != nil {
		_ = connection.Close()
		return
	}
	defer func() { _ = disposable.Dispose(); handler.sessions.Remove(connID); _ = connection.Close() }()
	authenticated := false
	authSignal := make(chan struct{})
	go func() {
		if waitErr := timeoutUseCase.Wait(ctx, connID, authSignal); waitErr != nil {
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
	machineID := ""
	for {
		messageType, payload, readErr := connection.ReadMessage()
		if readErr != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		frames, decodeErr := codec.DecodeFrames(payload)
		if decodeErr != nil {
			continue
		}
		for _, frame := range frames {
			handler.logger.Debug("websocket packet received", zap.String("conn_id", connID), zap.Uint16("packet_id", frame.PacketID), zap.Int("size", len(frame.Body)))
			if frame.PacketID == packetauth.ClientMachineIDPacketID {
				packet := packetauth.ClientMachineIDPacket{}
				if packet.Decode(frame.Body) == nil {
					machineID, _ = handler.policy.Normalize(packet.MachineID)
					response := packetauth.ServerMachineIDPacket{MachineID: machineID}
					body, bodyErr := response.Encode()
					if bodyErr == nil {
						_ = transport.Send(connID, response.PacketID(), body)
					}
				}
			}
			if frame.PacketID == packetauth.SSOTicketPacketID {
				packet := packetauth.SSOTicketPacket{}
				if packet.Decode(frame.Body) != nil {
					continue
				}
				_, authErr := useCase.Authenticate(ctx, authflow.AuthenticateRequest{ConnID: connID, Ticket: packet.Ticket, MachineID: machineID})
				if authErr == nil && !authenticated {
					authenticated = true
					close(authSignal)
				}
				if authErr != nil {
					return
				}
			}
		}
	}
}
