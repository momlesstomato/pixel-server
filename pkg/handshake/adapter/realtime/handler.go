package realtime

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/cryptoflow"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/sessionflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	packetsession "github.com/momlesstomato/pixel-server/pkg/handshake/packet/session"
	"go.uber.org/zap"
)

// Handler defines websocket handshake runtime behavior.
type Handler struct {
	// validator validates SSO tickets during authentication flow.
	validator authflow.TicketValidator
	// sessions stores connection lifecycle state in the session registry.
	sessions coreconnection.SessionRegistry
	// policy validates and regenerates machine identifiers.
	policy *packetauth.MachineIDPolicy
	// bus publishes and receives distributed close instructions.
	bus CloseSignalBus
	// logger stores runtime structured log behavior.
	logger *zap.Logger
	// authTimeout stores authentication timeout duration.
	authTimeout time.Duration
	// heartbeatInterval stores heartbeat ping interval.
	heartbeatInterval time.Duration
	// heartbeatTimeout stores heartbeat pong timeout.
	heartbeatTimeout time.Duration
	// connID creates stable connection identifiers.
	connID func() (string, error)
}

// runtimeUseCases defines handshake runtime use-case wiring behavior.
type runtimeUseCases struct {
	// authenticate stores SSO authentication workflow behavior.
	authenticate *authflow.AuthenticateUseCase
	// timeout stores unauthenticated timeout workflow behavior.
	timeout *authflow.TimeoutUseCase
	// disconnect stores disconnect workflow behavior.
	disconnect *sessionflow.DisconnectUseCase
	// heartbeat stores heartbeat ping/pong workflow behavior.
	heartbeat *sessionflow.HeartbeatUseCase
	// latency stores latency request/response workflow behavior.
	latency *sessionflow.LatencyUseCase
	// crypto stores diffie/rsa/rc4 exchange workflow behavior.
	crypto *cryptoflow.Session
}

// NewHandler creates websocket handshake runtime behavior.
func NewHandler(validator authflow.TicketValidator, sessions coreconnection.SessionRegistry, policy *packetauth.MachineIDPolicy, bus CloseSignalBus, logger *zap.Logger, authTimeout time.Duration) (*Handler, error) {
	return NewHandlerWithHeartbeat(validator, sessions, policy, bus, logger, authTimeout, 0, 0)
}

// NewHandlerWithHeartbeat creates websocket handshake runtime with heartbeat settings.
func NewHandlerWithHeartbeat(validator authflow.TicketValidator, sessions coreconnection.SessionRegistry, policy *packetauth.MachineIDPolicy, bus CloseSignalBus, logger *zap.Logger, authTimeout time.Duration, heartbeatInterval time.Duration, heartbeatTimeout time.Duration) (*Handler, error) {
	if validator == nil {
		return nil, fmt.Errorf("ticket validator is required")
	}
	if sessions == nil {
		return nil, fmt.Errorf("session registry is required")
	}
	if bus == nil {
		return nil, fmt.Errorf("close signal bus is required")
	}
	appliedPolicy := policy
	if appliedPolicy == nil {
		appliedPolicy = packetauth.NewMachineIDPolicy(nil)
	}
	output := logger
	if output == nil {
		output = zap.NewNop()
	}
	return &Handler{validator: validator, sessions: sessions, policy: appliedPolicy, bus: bus, logger: output, authTimeout: authTimeout, heartbeatInterval: heartbeatInterval, heartbeatTimeout: heartbeatTimeout, connID: func() (string, error) { return GenerateConnectionID(rand.Reader) }}, nil
}

// GenerateConnectionID creates one connection identifier string.
func GenerateConnectionID(source io.Reader) (string, error) {
	reader := source
	if reader == nil {
		reader = rand.Reader
	}
	buffer := make([]byte, 16)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return "", fmt.Errorf("generate connection id: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

// newRuntimeUseCases creates runtime handshake workflow dependencies for one connection.
func (handler *Handler) newRuntimeUseCases(transport *Transport) (*runtimeUseCases, error) {
	authenticate, authErr := authflow.NewAuthenticateUseCase(handler.validator, handler.sessions, transport)
	timeout, timeoutErr := authflow.NewTimeoutUseCase(transport, handler.authTimeout)
	disconnect, disconnectErr := sessionflow.NewDisconnectUseCase(handler.sessions, transport)
	heartbeat, heartbeatErr := sessionflow.NewHeartbeatUseCase(transport, handler.heartbeatInterval, handler.heartbeatTimeout)
	latency, latencyErr := sessionflow.NewLatencyUseCase(transport)
	crypto, cryptoErr := cryptoflow.NewSession(cryptoflow.Options{ServerClientEncryption: true})
	if authErr != nil || timeoutErr != nil || disconnectErr != nil || heartbeatErr != nil || latencyErr != nil || cryptoErr != nil {
		return nil, fmt.Errorf("handshake runtime initialization failed")
	}
	return &runtimeUseCases{authenticate: authenticate, timeout: timeout, disconnect: disconnect, heartbeat: heartbeat, latency: latency, crypto: crypto}, nil
}

// handleMachineID normalizes and echoes machine identifier payload.
func (handler *Handler) handleMachineID(connID string, body []byte, transport *Transport, machineID *string) {
	packet := packetauth.ClientMachineIDPacket{}
	if packet.Decode(body) != nil {
		return
	}
	normalized, err := handler.policy.Normalize(packet.MachineID)
	if err != nil {
		return
	}
	*machineID = normalized
	response := packetauth.ServerMachineIDPacket{MachineID: normalized}
	encoded, encodeErr := response.Encode()
	if encodeErr == nil {
		_ = transport.Send(connID, response.PacketID(), encoded)
	}
}

// handleSSO authenticates one SSO packet payload.
func (handler *Handler) handleSSO(ctx context.Context, connID string, body []byte, machineID string, useCase *authflow.AuthenticateUseCase) bool {
	packet := packetauth.SSOTicketPacket{}
	if packet.Decode(body) != nil {
		return true
	}
	_, err := useCase.Authenticate(ctx, authflow.AuthenticateRequest{ConnID: connID, Ticket: packet.Ticket, MachineID: machineID})
	return err == nil
}

// handleDisconnect handles one client disconnect packet and closes connection.
func (handler *Handler) handleDisconnect(connID string, body []byte, useCase *sessionflow.DisconnectUseCase) bool {
	packet := packetsession.ClientDisconnectPacket{}
	if packet.Decode(body) != nil {
		return false
	}
	_ = useCase.Disconnect(connID)
	return true
}
