package app

import (
	sessionmessaging "pixelsv/internal/sessionconnection/messaging"
	"pixelsv/pkg/plugin"
)

// RecordPacket emits one packet-received plugin event for a session packet.
func (s *Service) RecordPacket(sessionID string, header uint16, packetName string) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	s.mu.Lock()
	_, ok := s.sessions[sessionID]
	s.mu.Unlock()
	if !ok {
		return ErrSessionNotFound
	}
	s.emit(&plugin.Event{
		Name:      sessionmessaging.EventPacketReceived,
		SessionID: sessionID,
		Data: sessionmessaging.PacketReceivedEventData{
			SessionID:  sessionID,
			Header:     header,
			PacketName: packetName,
		},
	})
	return nil
}
