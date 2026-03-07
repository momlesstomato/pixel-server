package app

import (
	"strings"
	"time"
)

// handshakeSession stores per-session handshake state.
type handshakeSession struct {
	// CreatedAt is the first observed packet timestamp.
	CreatedAt time.Time
	// UpdatedAt is the latest observed packet timestamp.
	UpdatedAt time.Time
	// ReleaseVersion stores reported client release.
	ReleaseVersion string
	// ClientType stores reported client type string.
	ClientType string
	// Platform stores reported platform identifier.
	Platform int32
	// DeviceCategory stores reported device category.
	DeviceCategory int32
	// ClientID stores reported client variables id.
	ClientID int32
	// ClientURL stores reported client URL.
	ClientURL string
	// ExternalVariablesURL stores reported external variables URL.
	ExternalVariablesURL string
	// MachineID stores normalized machine id.
	MachineID string
	// Fingerprint stores reported device fingerprint.
	Fingerprint string
	// Capabilities stores reported client capabilities string.
	Capabilities string
	// DiffieState stores active diffie session values.
	DiffieState *diffieSession
	// DiffieComplete indicates complete_diffie was processed.
	DiffieComplete bool
	// Authenticated indicates successful sso validation.
	Authenticated bool
}

// RemoveSession removes one session handshake state.
func (s *Service) RemoveSession(sessionID string) {
	if strings.TrimSpace(sessionID) == "" {
		return
	}
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

// ExpireUnauthenticatedSessions removes expired unauthenticated sessions.
func (s *Service) ExpireUnauthenticatedSessions(now time.Time) []string {
	expired := make([]string, 0, 8)
	s.mu.Lock()
	for sessionID, session := range s.sessions {
		if session.Authenticated {
			continue
		}
		if now.Sub(session.CreatedAt) >= s.timeout {
			expired = append(expired, sessionID)
			delete(s.sessions, sessionID)
		}
	}
	s.mu.Unlock()
	return expired
}

// touchSession returns mutable session state, creating it when missing.
func (s *Service) touchSession(sessionID string) (*handshakeSession, error) {
	clean := strings.TrimSpace(sessionID)
	if clean == "" {
		return nil, ErrInvalidSessionID
	}
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[clean]
	if !ok {
		session = &handshakeSession{CreatedAt: now, UpdatedAt: now}
		s.sessions[clean] = session
		return session, nil
	}
	session.UpdatedAt = now
	return session, nil
}
