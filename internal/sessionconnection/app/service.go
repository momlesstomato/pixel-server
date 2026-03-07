package app

import (
	"errors"
	"sort"
	"sync"
	"time"

	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/plugin"
)

var (
	// ErrInvalidSessionID indicates session operations require non-empty session id.
	ErrInvalidSessionID = errors.New("invalid session id")
	// ErrSessionNotFound indicates a session operation referenced unknown session state.
	ErrSessionNotFound = errors.New("session not found")
)

// Service orchestrates session-connection use cases.
type Service struct {
	// events emits session lifecycle events for plugin hooks.
	events plugin.EventBus
	// telemetryMinInterval limits frequency of telemetry packet logging.
	telemetryMinInterval time.Duration
	// sessions stores session runtime state by session id.
	sessions map[string]*sessionState
	// users maps user id to active session id for concurrent-login enforcement.
	users map[int32]string
	// mu serializes all state updates.
	mu sync.Mutex
	// clock resolves current time and supports deterministic tests.
	clock func() time.Time
}

// NewService creates a new session-connection application service.
func NewService(events plugin.EventBus, telemetryMinInterval time.Duration) *Service {
	return &Service{
		events:               events,
		telemetryMinInterval: telemetryMinInterval,
		sessions:             map[string]*sessionState{},
		users:                map[int32]string{},
		clock:                time.Now,
	}
}

// ActiveAuthenticatedSessions returns sorted authenticated session ids.
func (s *Service) ActiveAuthenticatedSessions() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.sessions))
	for sessionID, state := range s.sessions {
		if state.Authenticated {
			ids = append(ids, sessionID)
		}
	}
	sort.Strings(ids)
	return ids
}

// ExpirePongTimeoutSessions removes stale authenticated sessions and returns ids.
func (s *Service) ExpirePongTimeoutSessions(timeout time.Duration, now time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	expired := make([]string, 0, len(s.sessions))
	for sessionID, state := range s.sessions {
		if !state.Authenticated || now.Sub(state.LastPongAt) <= timeout {
			continue
		}
		delete(s.users, state.UserID)
		delete(s.sessions, sessionID)
		expired = append(expired, sessionID)
	}
	sort.Strings(expired)
	return expired
}

// emit sends one plugin event when an event bus is available.
func (s *Service) emit(event *plugin.Event) {
	if s.events != nil {
		_ = s.events.Emit(event)
	}
}

// validateSessionID validates that one session id value is not empty.
func validateSessionID(sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSessionID
	}
	return nil
}

// connectedEvent builds one connected plugin event payload.
func connectedEvent(sessionID string) *plugin.Event {
	return &plugin.Event{Name: sessionmessaging.EventSessionConnected, SessionID: sessionID, Data: sessionmessaging.SessionConnectedEventData{SessionID: sessionID}}
}
