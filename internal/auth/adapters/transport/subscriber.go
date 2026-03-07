package transport

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"go.uber.org/zap"
	"pixelsv/internal/auth/app"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/protocol"
)

// Service defines auth application behavior consumed by transport adapter.
type Service interface {
	// RecordReleaseVersion stores release metadata for one session.
	RecordReleaseVersion(sessionID string, packet *protocol.HandshakeReleaseVersionPacket) error
	// RecordClientVariables stores client metadata for one session.
	RecordClientVariables(sessionID string, packet *protocol.HandshakeClientVariablesPacket) error
	// InitDiffie initializes diffie values and returns response fields.
	InitDiffie(sessionID string) (app.InitDiffieResponse, error)
	// CompleteDiffie finalizes diffie values and returns response fields.
	CompleteDiffie(sessionID string, encryptedPublicKey string) (app.CompleteDiffieResponse, error)
	// UpdateMachineID stores machine metadata and returns normalization details.
	UpdateMachineID(sessionID string, machineID string, fingerprint string, capabilities string) (string, bool, error)
	// MarkLatencyMeasure tracks latency packet receipt.
	MarkLatencyMeasure(sessionID string) error
	// MarkClientPolicy tracks client policy packet receipt.
	MarkClientPolicy(sessionID string) error
	// ValidateTicket validates and consumes one SSO ticket.
	ValidateTicket(sessionID string, ticket string) (int32, error)
	// RemoveSession removes one session handshake state.
	RemoveSession(sessionID string)
	// ExpireUnauthenticatedSessions evicts expired handshake sessions.
	ExpireUnauthenticatedSessions(now time.Time) []string
}

// Subscriber consumes handshake packet topics and emits auth events.
type Subscriber struct {
	// bus is the runtime transport bus.
	bus coretransport.Bus
	// service provides ticket and handshake use cases.
	service Service
	// logger stores adapter logs.
	logger *zap.Logger
}

// NewSubscriber creates a new auth transport subscriber.
func NewSubscriber(bus coretransport.Bus, service Service, logger *zap.Logger) *Subscriber {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Subscriber{bus: bus, service: service, logger: logger}
}

// Start subscribes handshake-security packet ingress and lifecycle topics.
func (s *Subscriber) Start(ctx context.Context) error {
	if _, err := s.bus.Subscribe(ctx, authmessaging.PacketIngressWildcardTopic(), s.handlePacket); err != nil {
		return err
	}
	if _, err := s.bus.Subscribe(ctx, sessionmessaging.TopicDisconnected, s.handleSessionDisconnected); err != nil {
		return err
	}
	go s.monitorHandshakeTimeouts(ctx)
	return nil
}

// handlePacket decodes one ingress packet and routes known actions.
func (s *Subscriber) handlePacket(ctx context.Context, message coretransport.Message) error {
	sessionID, ok := authmessaging.ParsePacketIngressTopic(message.Topic)
	if !ok {
		return nil
	}
	packet, err := decodePacket(message.Payload)
	if err != nil || packet == nil {
		return err
	}
	s.logger.Debug(
		"auth packet received",
		zap.String("session_id", sessionID),
		zap.Uint16("header", packet.HeaderID()),
		zap.String("packet", packet.PacketName()),
	)
	return s.dispatchPacket(ctx, sessionID, packet)
}

// handleSessionDisconnected clears auth session state on gateway disconnect events.
func (s *Subscriber) handleSessionDisconnected(ctx context.Context, message coretransport.Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	sessionID := string(message.Payload)
	if sessionID != "" {
		s.service.RemoveSession(sessionID)
	}
	return nil
}

// monitorHandshakeTimeouts disconnects sessions that do not authenticate in time.
func (s *Subscriber) monitorHandshakeTimeouts(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.handleExpiredSessions(ctx, now)
		}
	}
}

// decodePacket decodes one packet payload body into a protocol packet.
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
