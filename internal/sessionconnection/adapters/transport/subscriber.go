package transport

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"go.uber.org/zap"
	"pixelsv/internal/sessionconnection/app"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/protocol"
)

// Config defines transport subscriber runtime settings.
type Config struct {
	// PingInterval defines keepalive ping interval.
	PingInterval time.Duration
	// PongTimeout defines max pong silence before disconnect.
	PongTimeout time.Duration
	// AvailabilityOpen controls availability.status isOpen flag.
	AvailabilityOpen bool
	// AvailabilityOnShutdown controls availability.status onShutdown flag.
	AvailabilityOnShutdown bool
	// AvailabilityAuthentic controls availability.status isAuthentic flag.
	AvailabilityAuthentic bool
}

// Service defines session-connection app behavior consumed by transport adapter.
type Service interface {
	// SessionConnected initializes state for one session id.
	SessionConnected(sessionID string) error
	// SessionDisconnected removes one session state.
	SessionDisconnected(sessionID string)
	// SessionAuthenticated marks one session authenticated and returns previous session id.
	SessionAuthenticated(sessionID string, userID int32) (string, error)
	// MarkPong updates one session pong timestamp.
	MarkPong(sessionID string) error
	// MarkLatencyTest records one session latency-test packet.
	MarkLatencyTest(sessionID string, requestID int32) error
	// MarkDesktopView records one desktop-view signal.
	MarkDesktopView(sessionID string) error
	// RecordPacket emits packet-level plugin event metadata.
	RecordPacket(sessionID string, header uint16, packetName string) error
	// AllowTelemetry applies telemetry packet log throttling.
	AllowTelemetry(sessionID string, header uint16) (bool, error)
	// ActiveAuthenticatedSessions returns authenticated session ids.
	ActiveAuthenticatedSessions() []string
	// ExpirePongTimeoutSessions removes stale sessions and returns ids.
	ExpirePongTimeoutSessions(timeout time.Duration, now time.Time) []string
}

// Subscriber handles session-connection transport wiring and packet behavior.
type Subscriber struct {
	// bus is the runtime transport bus.
	bus coretransport.Bus
	// service provides session-connection application behavior.
	service Service
	// logger stores adapter logs.
	logger *zap.Logger
	// cfg stores keepalive and availability behavior.
	cfg Config
}

// NewSubscriber creates a new session-connection transport subscriber.
func NewSubscriber(bus coretransport.Bus, service Service, logger *zap.Logger, cfg Config) *Subscriber {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Subscriber{bus: bus, service: service, logger: logger, cfg: cfg}
}

// Start subscribes transport topics and starts keepalive routines.
func (s *Subscriber) Start(ctx context.Context) error {
	subscriptionsCtx := context.Background()
	if _, err := s.bus.Subscribe(subscriptionsCtx, sessionmessaging.TopicConnected, s.handleConnected); err != nil {
		return err
	}
	if _, err := s.bus.Subscribe(subscriptionsCtx, sessionmessaging.TopicDisconnected, s.handleDisconnected); err != nil {
		return err
	}
	if _, err := s.bus.Subscribe(subscriptionsCtx, sessionmessaging.TopicAuthenticated, s.handleAuthenticated); err != nil {
		return err
	}
	if _, err := s.bus.Subscribe(subscriptionsCtx, sessionmessaging.PacketIngressWildcardTopic(), s.handlePacket); err != nil {
		return err
	}
	go s.runPingLoop(ctx)
	go s.runTimeoutLoop(ctx)
	return nil
}

// decodePacket decodes one ingress payload body into a session-connection packet.
func decodePacket(body []byte) (protocol.Packet, error) {
	if len(body) < 2 {
		return nil, codec.ErrInvalidFrame
	}
	header := binary.BigEndian.Uint16(body[:2])
	packet, err := protocol.DecodeC2S(header, body[2:])
	if err != nil {
		if errors.Is(err, protocol.ErrUnknownHeader) {
			return nil, nil
		}
		return nil, err
	}
	return packet, nil
}

// ignoreMissing suppresses session-not-found errors for stale lifecycle races.
func ignoreMissing(err error) error {
	if errors.Is(err, app.ErrSessionNotFound) {
		return nil
	}
	return err
}
