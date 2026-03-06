package transport

import (
	"context"
	"encoding/binary"
	"errors"

	"go.uber.org/zap"
	"pixelsv/internal/auth/app"
	"pixelsv/internal/auth/domain"
	authmessaging "pixelsv/internal/auth/messaging"
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/codec"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/protocol"
)

// Service defines auth application behavior consumed by transport adapter.
type Service interface {
	// ValidateTicket validates and consumes one SSO ticket.
	ValidateTicket(ticket string) (int32, error)
}

// Subscriber consumes handshake packet topics and emits auth events.
type Subscriber struct {
	// bus is the runtime transport bus.
	bus coretransport.Bus
	// service provides ticket validation use cases.
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

// Start subscribes handshake-security packet ingress topics.
func (s *Subscriber) Start(ctx context.Context) error {
	_, err := s.bus.Subscribe(ctx, authmessaging.PacketIngressWildcardTopic(), s.handlePacket)
	return err
}

func (s *Subscriber) handlePacket(ctx context.Context, message coretransport.Message) error {
	sessionID, ok := authmessaging.ParsePacketIngressTopic(message.Topic)
	if !ok {
		return nil
	}
	packet, err := decodePacket(message.Payload)
	if err != nil || packet == nil {
		return err
	}
	switch value := packet.(type) {
	case *protocol.SecuritySsoTicketPacket:
		return s.handleSSOTicket(ctx, sessionID, value.Ticket)
	default:
		return nil
	}
}

func (s *Subscriber) handleSSOTicket(ctx context.Context, sessionID string, ticket string) error {
	userID, err := s.service.ValidateTicket(ticket)
	if err != nil {
		if errors.Is(err, domain.ErrTicketNotFound) || errors.Is(err, domain.ErrInvalidTicket) {
			s.logger.Warn("ticket validation failed", zap.String("session_id", sessionID), zap.Error(err))
			return nil
		}
		return err
	}
	authenticated := app.EncodeAuthenticatedEvent(sessionID, userID)
	if err := s.bus.Publish(ctx, sessionmessaging.TopicAuthenticated, authenticated); err != nil {
		return err
	}
	return s.bus.Publish(ctx, sessionmessaging.OutputTopic(sessionID), codec.EncodeFrame(2491, nil))
}

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
