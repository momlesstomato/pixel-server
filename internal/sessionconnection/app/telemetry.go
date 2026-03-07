package app

import "time"

// AllowTelemetry reports whether telemetry packet should be logged now.
func (s *Service) AllowTelemetry(sessionID string, header uint16) (bool, error) {
	if err := validateSessionID(sessionID); err != nil {
		return false, err
	}
	now := s.clock()
	s.mu.Lock()
	state, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return false, ErrSessionNotFound
	}
	last := state.LastTelemetryByHeader[header]
	if !last.IsZero() && now.Sub(last) < s.telemetryMinInterval {
		s.mu.Unlock()
		return false, nil
	}
	state.LastTelemetryByHeader[header] = now
	s.mu.Unlock()
	return true, nil
}

// sessionState stores mutable runtime session-connection state.
type sessionState struct {
	// SessionID is the runtime websocket session identifier.
	SessionID string
	// UserID is authenticated user identifier.
	UserID int32
	// Authenticated reports whether auth has validated this session.
	Authenticated bool
	// LastPongAt stores last observed client pong timestamp.
	LastPongAt time.Time
	// LastTelemetryByHeader tracks telemetry throttling by packet header.
	LastTelemetryByHeader map[uint16]time.Time
}
