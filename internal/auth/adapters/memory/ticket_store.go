package memory

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"pixelsv/internal/auth/domain"
)

type record struct {
	userID    int32
	expiresAt time.Time
}

// TicketStore stores SSO tickets in memory with TTL expiration checks.
type TicketStore struct {
	mu      sync.Mutex
	tickets map[string]record
}

// NewTicketStore creates an empty in-memory ticket store.
func NewTicketStore() *TicketStore {
	return &TicketStore{tickets: map[string]record{}}
}

// Create generates one ticket for a user id.
func (s *TicketStore) Create(userID int32, ttl time.Duration) (string, error) {
	if userID <= 0 {
		return "", domain.ErrInvalidUserID
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.tickets[token] = record{userID: userID, expiresAt: time.Now().Add(ttl)}
	s.mu.Unlock()
	return token, nil
}

// Consume validates and removes one ticket.
func (s *TicketStore) Consume(ticket string) (int32, error) {
	if ticket == "" {
		return 0, domain.ErrInvalidTicket
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.tickets[ticket]
	if !ok || time.Now().After(value.expiresAt) {
		delete(s.tickets, ticket)
		return 0, domain.ErrTicketNotFound
	}
	delete(s.tickets, ticket)
	return value.userID, nil
}

// Revoke removes one ticket.
func (s *TicketStore) Revoke(ticket string) error {
	if ticket == "" {
		return domain.ErrInvalidTicket
	}
	s.mu.Lock()
	delete(s.tickets, ticket)
	s.mu.Unlock()
	return nil
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
