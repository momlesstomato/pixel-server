package app

import (
	"strings"

	authmessaging "pixelsv/internal/auth/messaging"
	"pixelsv/pkg/plugin"
	"pixelsv/pkg/protocol"
)

// InitDiffieResponse carries init_diffie server response fields.
type InitDiffieResponse struct {
	// SignedPrime is the RSA-signed prime value.
	SignedPrime string
	// SignedGenerator is the RSA-signed generator value.
	SignedGenerator string
}

// CompleteDiffieResponse carries complete_diffie server response fields.
type CompleteDiffieResponse struct {
	// PublicKey is the server public key value.
	PublicKey string
	// ServerEncryption indicates whether encryption is active.
	ServerEncryption bool
}

// RecordReleaseVersion stores release metadata for one handshake session.
func (s *Service) RecordReleaseVersion(sessionID string, packet *protocol.HandshakeReleaseVersionPacket) error {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return err
	}
	if packet == nil || strings.TrimSpace(packet.ReleaseVersion) == "" {
		return ErrReleaseVersionRequired
	}
	if len(s.releaseAllowlist) > 0 {
		if _, ok := s.releaseAllowlist[packet.ReleaseVersion]; !ok {
			return ErrUnsupportedReleaseVersion
		}
	}
	s.mu.Lock()
	state.ReleaseVersion = packet.ReleaseVersion
	state.ClientType = packet.ClientType
	state.Platform = packet.Platform
	state.DeviceCategory = packet.DeviceCategory
	s.mu.Unlock()
	if s.events != nil {
		_ = s.events.Emit(&plugin.Event{
			Name:      authmessaging.EventReleaseVersionReceived,
			SessionID: sessionID,
			Data: authmessaging.ReleaseVersionEventData{
				ReleaseVersion: packet.ReleaseVersion,
				ClientType:     packet.ClientType,
				Platform:       packet.Platform,
				DeviceCategory: packet.DeviceCategory,
			},
		})
	}
	return nil
}

// RecordClientVariables stores client variables packet data for diagnostics.
func (s *Service) RecordClientVariables(sessionID string, packet *protocol.HandshakeClientVariablesPacket) error {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return err
	}
	if packet == nil {
		return nil
	}
	s.mu.Lock()
	state.ClientID = packet.ClientId
	state.ClientURL = packet.ClientUrl
	state.ExternalVariablesURL = packet.ExternalVariablesUrl
	s.mu.Unlock()
	return nil
}

// InitDiffie initializes diffie state and returns init_diffie response values.
func (s *Service) InitDiffie(sessionID string) (InitDiffieResponse, error) {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return InitDiffieResponse{}, err
	}
	session, signedPrime, signedGenerator, err := s.crypto.startDiffie()
	if err != nil {
		return InitDiffieResponse{}, err
	}
	s.mu.Lock()
	state.DiffieState = session
	state.DiffieComplete = false
	s.mu.Unlock()
	if s.events != nil {
		_ = s.events.Emit(&plugin.Event{
			Name:      authmessaging.EventDiffieInitialized,
			SessionID: sessionID,
			Data:      authmessaging.DiffieInitializedEventData{SessionID: sessionID},
		})
	}
	return InitDiffieResponse{SignedPrime: signedPrime, SignedGenerator: signedGenerator}, nil
}

// CompleteDiffie completes diffie exchange and returns server response values.
func (s *Service) CompleteDiffie(sessionID string, encryptedPublicKey string) (CompleteDiffieResponse, error) {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return CompleteDiffieResponse{}, err
	}
	publicKey, err := s.crypto.completeDiffie(state.DiffieState, encryptedPublicKey)
	if err != nil {
		return CompleteDiffieResponse{}, err
	}
	s.mu.Lock()
	state.DiffieComplete = true
	s.mu.Unlock()
	if s.events != nil {
		_ = s.events.Emit(&plugin.Event{
			Name:      authmessaging.EventDiffieCompleted,
			SessionID: sessionID,
			Data:      authmessaging.DiffieCompletedEventData{SessionID: sessionID},
		})
	}
	return CompleteDiffieResponse{PublicKey: publicKey, ServerEncryption: false}, nil
}
