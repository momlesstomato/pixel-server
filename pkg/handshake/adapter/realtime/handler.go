package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/cryptoflow"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/sessionflow"
	packetauth "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	sessionnavigation "github.com/momlesstomato/pixel-server/pkg/session/application/navigation"
	sessionpostauth "github.com/momlesstomato/pixel-server/pkg/session/application/postauth"
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
	// broadcaster publishes and subscribes session notification channels.
	broadcaster broadcast.Broadcaster
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
	// postAuthFactory creates post-authentication burst behavior.
	postAuthFactory func(*Transport) (*sessionpostauth.UseCase, error)
	// desktopFactory creates desktop-view navigation behavior.
	desktopFactory func(*Transport) (*sessionnavigation.DesktopViewUseCase, error)
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
	// postauth stores post-authentication burst workflow behavior.
	postauth *sessionpostauth.UseCase
	// desktop stores desktop-view navigation workflow behavior.
	desktop *sessionnavigation.DesktopViewUseCase
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
	factory := func(_ *Transport) (*sessionpostauth.UseCase, error) { return nil, nil }
	desktopFactory := func(_ *Transport) (*sessionnavigation.DesktopViewUseCase, error) { return nil, nil }
	return &Handler{
		validator: validator, sessions: sessions, policy: appliedPolicy, bus: bus, logger: output,
		authTimeout: authTimeout, heartbeatInterval: heartbeatInterval, heartbeatTimeout: heartbeatTimeout,
		connID: func() (string, error) { return GenerateConnectionID(rand.Reader) }, postAuthFactory: factory, desktopFactory: desktopFactory,
	}, nil
}

// ConfigurePostAuth wires post-authentication packet burst dependencies.
func (handler *Handler) ConfigurePostAuth(status sessionpostauth.StatusReader, logins sessionpostauth.LoginRecorder, holder string) {
	handler.postAuthFactory = func(transport *Transport) (*sessionpostauth.UseCase, error) {
		return sessionpostauth.NewUseCase(transport, status, logins, holder)
	}
}

// ConfigureBroadcaster wires distributed broadcast channels for session notifications.
func (handler *Handler) ConfigureBroadcaster(broadcaster broadcast.Broadcaster) {
	handler.broadcaster = broadcaster
}

// ConfigureDesktopView wires desktop-view navigation behavior.
func (handler *Handler) ConfigureDesktopView(checker sessionnavigation.RoomChecker) {
	handler.desktopFactory = func(transport *Transport) (*sessionnavigation.DesktopViewUseCase, error) {
		return sessionnavigation.NewDesktopViewUseCase(transport, checker)
	}
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
	postauth, postAuthErr := handler.postAuthFactory(transport)
	desktop, desktopErr := handler.desktopFactory(transport)
	if authErr != nil || timeoutErr != nil || disconnectErr != nil || heartbeatErr != nil || latencyErr != nil || cryptoErr != nil || postAuthErr != nil || desktopErr != nil {
		return nil, fmt.Errorf("handshake runtime initialization failed")
	}
	return &runtimeUseCases{
		authenticate: authenticate, timeout: timeout, disconnect: disconnect, heartbeat: heartbeat,
		latency: latency, crypto: crypto, postauth: postauth, desktop: desktop,
	}, nil
}
