package app

import (
	"crypto/rand"
	"encoding/hex"
	"strings"

	authmessaging "pixelsv/internal/auth/messaging"
	"pixelsv/pkg/plugin"
)

// UpdateMachineID stores machine identifiers and returns normalization details.
func (s *Service) UpdateMachineID(sessionID string, machineID string, fingerprint string, capabilities string) (string, bool, error) {
	state, err := s.touchSession(sessionID)
	if err != nil {
		return "", false, err
	}
	normalized := machineID
	changed := false
	if strings.HasPrefix(normalized, "~") || len(normalized) != 64 {
		normalized = randomHexString(64)
		changed = true
	}
	s.mu.Lock()
	state.MachineID = normalized
	state.Fingerprint = fingerprint
	state.Capabilities = capabilities
	s.mu.Unlock()
	if s.events != nil {
		_ = s.events.Emit(&plugin.Event{
			Name:      authmessaging.EventMachineIDReceived,
			SessionID: sessionID,
			Data:      authmessaging.MachineIDEventData{MachineID: normalized, Changed: changed},
		})
	}
	return normalized, changed, nil
}

// MarkLatencyMeasure records receipt of latency packet.
func (s *Service) MarkLatencyMeasure(sessionID string) error {
	_, err := s.touchSession(sessionID)
	return err
}

// MarkClientPolicy records receipt of client policy packet.
func (s *Service) MarkClientPolicy(sessionID string) error {
	_, err := s.touchSession(sessionID)
	return err
}

// randomHexString returns a random lowercase hex value with exact length.
func randomHexString(length int) string {
	if length <= 0 {
		return ""
	}
	buffer := make([]byte, (length+1)/2)
	if _, err := rand.Read(buffer); err != nil {
		return strings.Repeat("0", length)
	}
	value := hex.EncodeToString(buffer)
	if len(value) >= length {
		return value[:length]
	}
	return value + strings.Repeat("0", length-len(value))
}
