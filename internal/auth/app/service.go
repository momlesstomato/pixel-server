package app

import (
	"errors"
	"sync"
	"time"

	"pixelsv/internal/auth/domain"
	authmessaging "pixelsv/internal/auth/messaging"
	"pixelsv/pkg/plugin"
)

var (
	// ErrInvalidSessionID indicates handshake operations require a session id.
	ErrInvalidSessionID = errors.New("invalid session id")
	// ErrReleaseVersionRequired indicates authentication requires release version first.
	ErrReleaseVersionRequired = errors.New("release version is required before authentication")
	// ErrUnsupportedReleaseVersion indicates release version is not in allowlist.
	ErrUnsupportedReleaseVersion = errors.New("unsupported release version")
	// ErrDiffieNotInitialized indicates complete-diffie requires init-diffie state.
	ErrDiffieNotInitialized = errors.New("diffie handshake is not initialized")
	// ErrDiffieRequired indicates authentication requires completed diffie handshake.
	ErrDiffieRequired = errors.New("diffie handshake must complete before authentication")
)

// Service orchestrates auth realm ticket and handshake use cases.
type Service struct {
	// store provides ticket persistence behavior.
	store domain.TicketStore
	// events emits realm lifecycle events for plugin hooks.
	events plugin.EventBus
	// crypto holds cryptographic handshake primitives.
	crypto *handshakeCrypto
	// sessions stores ephemeral per-session handshake state.
	sessions map[string]*handshakeSession
	// releaseAllowlist constrains accepted client releases when non-empty.
	releaseAllowlist map[string]struct{}
	// timeout defines unauthenticated handshake expiration.
	timeout time.Duration
	// requireDiffie enforces diffie completion before authentication.
	requireDiffie bool
	// mu protects session state mutations.
	mu sync.Mutex
	// clock resolves current time and supports deterministic tests.
	clock func() time.Time
}

// NewService creates a new auth application service.
func NewService(store domain.TicketStore, events plugin.EventBus) *Service {
	return &Service{
		store:            store,
		events:           events,
		crypto:           newHandshakeCrypto(),
		sessions:         make(map[string]*handshakeSession),
		releaseAllowlist: map[string]struct{}{},
		timeout:          15 * time.Second,
		requireDiffie:    false,
		clock:            time.Now,
	}
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

// ValidateTicket validates and consumes one ticket for a session.
func (s *Service) ValidateTicket(sessionID string, ticket string) (int32, error) {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return 0, err
	}
	if state.ReleaseVersion == "" {
		return 0, ErrReleaseVersionRequired
	}
	if s.requireDiffie && !state.DiffieComplete {
		return 0, ErrDiffieRequired
	}
	userID, err := s.store.Consume(ticket)
	if err != nil {
		return 0, err
	}
	s.mu.Lock()
	state.Authenticated = true
	state.UpdatedAt = s.clock()
	s.mu.Unlock()
	if s.events != nil {
		_ = s.events.Emit(&plugin.Event{
			Name:      authmessaging.EventTicketValidated,
			SessionID: sessionID,
			Data:      authmessaging.TicketValidatedEventData{UserID: userID},
		})
	}
	return userID, nil
}

// RevokeTicket revokes one ticket.
func (s *Service) RevokeTicket(ticket string) error {
	return s.store.Revoke(ticket)
}
