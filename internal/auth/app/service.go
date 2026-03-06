package app

import (
	"time"

	"pixelsv/internal/auth/domain"
	"pixelsv/pkg/codec"
)

// Service orchestrates auth realm ticket use cases.
type Service struct {
	// store provides ticket persistence behavior.
	store domain.TicketStore
}

// NewService creates a new auth application service.
func NewService(store domain.TicketStore) *Service {
	return &Service{store: store}
}

// CreateTicket creates one ticket for a user with optional ttl seconds.
func (s *Service) CreateTicket(userID int32, ttlSeconds int32) (string, int32, error) {
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttlSeconds <= 0 {
		ttlSeconds = 300
		ttl = 300 * time.Second
	}
	ticket, err := s.store.Create(userID, ttl)
	if err != nil {
		return "", 0, err
	}
	return ticket, ttlSeconds, nil
}

// ValidateTicket validates and consumes one ticket.
func (s *Service) ValidateTicket(ticket string) (int32, error) {
	return s.store.Consume(ticket)
}

// RevokeTicket revokes one ticket.
func (s *Service) RevokeTicket(ticket string) error {
	return s.store.Revoke(ticket)
}

// EncodeAuthenticatedEvent encodes one session authenticated event payload.
func EncodeAuthenticatedEvent(sessionID string, userID int32) []byte {
	writer := codec.NewWriter(32)
	writer.WriteString(sessionID)
	writer.WriteInt32(userID)
	return writer.Bytes()
}
