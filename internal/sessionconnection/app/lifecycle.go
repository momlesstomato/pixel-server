package app

import (
	"time"

	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/plugin"
)

// SessionConnected initializes one runtime session state.
func (s *Service) SessionConnected(sessionID string) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	now := s.clock()
	s.mu.Lock()
	s.sessions[sessionID] = &sessionState{SessionID: sessionID, LastPongAt: now, LastTelemetryByHeader: map[uint16]time.Time{}}
	s.mu.Unlock()
	s.emit(connectedEvent(sessionID))
	return nil
}

// SessionDisconnected removes one runtime session state.
func (s *Service) SessionDisconnected(sessionID string) {
	if sessionID == "" {
		return
	}
	s.mu.Lock()
	state, ok := s.sessions[sessionID]
	if ok {
		delete(s.users, state.UserID)
		delete(s.sessions, sessionID)
	}
	s.mu.Unlock()
	if ok {
		s.emit(&plugin.Event{
			Name:      sessionmessaging.EventSessionDisconnected,
			SessionID: sessionID,
			Data:      sessionmessaging.SessionDisconnectedEventData{SessionID: sessionID},
		})
	}
}

// SessionAuthenticated marks one session as authenticated and enforces single-login.
func (s *Service) SessionAuthenticated(sessionID string, userID int32) (string, error) {
	if err := validateSessionID(sessionID); err != nil {
		return "", err
	}
	now := s.clock()
	s.mu.Lock()
	state, ok := s.sessions[sessionID]
	if !ok {
		s.mu.Unlock()
		return "", ErrSessionNotFound
	}
	previous := s.users[userID]
	state.Authenticated = true
	state.UserID = userID
	state.LastPongAt = now
	s.users[userID] = sessionID
	if previous != "" && previous != sessionID {
		delete(s.sessions, previous)
	}
	s.mu.Unlock()
	s.emit(&plugin.Event{
		Name:      sessionmessaging.EventSessionAuthenticated,
		SessionID: sessionID,
		Data:      sessionmessaging.SessionAuthenticatedEventData{SessionID: sessionID, UserID: userID},
	})
	return previous, nil
}

// MarkPong updates last pong timestamp for one session.
func (s *Service) MarkPong(sessionID string) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	s.mu.Lock()
	state, ok := s.sessions[sessionID]
	if ok {
		state.LastPongAt = s.clock()
	}
	s.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}
	s.emit(&plugin.Event{Name: sessionmessaging.EventClientPongReceived, SessionID: sessionID, Data: sessionmessaging.ClientPongEventData{SessionID: sessionID}})
	return nil
}

// MarkLatencyTest records one latency test request for a session.
func (s *Service) MarkLatencyTest(sessionID string, requestID int32) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	s.mu.Lock()
	_, ok := s.sessions[sessionID]
	s.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}
	s.emit(&plugin.Event{Name: sessionmessaging.EventLatencyTestReceived, SessionID: sessionID, Data: sessionmessaging.LatencyTestEventData{SessionID: sessionID, RequestID: requestID}})
	return nil
}

// MarkDesktopView records one desktop view signal for a session.
func (s *Service) MarkDesktopView(sessionID string) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	s.mu.Lock()
	_, ok := s.sessions[sessionID]
	s.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}
	s.emit(&plugin.Event{Name: sessionmessaging.EventDesktopViewReceived, SessionID: sessionID, Data: sessionmessaging.DesktopViewEventData{SessionID: sessionID}})
	return nil
}
