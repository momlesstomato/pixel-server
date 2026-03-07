package messaging

import (
	"errors"

	"pixelsv/pkg/codec"
)

var (
	// ErrInvalidAuthenticatedPayload indicates an authenticated payload is malformed.
	ErrInvalidAuthenticatedPayload = errors.New("invalid authenticated payload")
)

// AuthenticatedEvent stores one authenticated session event payload.
type AuthenticatedEvent struct {
	// SessionID is the runtime websocket session id.
	SessionID string
	// UserID is the authenticated user id.
	UserID int32
}

// EncodeAuthenticatedEvent encodes one authenticated event payload.
func EncodeAuthenticatedEvent(sessionID string, userID int32) []byte {
	writer := codec.NewWriter(32)
	writer.WriteString(sessionID)
	writer.WriteInt32(userID)
	return writer.Bytes()
}

// DecodeAuthenticatedEvent decodes one authenticated event payload.
func DecodeAuthenticatedEvent(payload []byte) (AuthenticatedEvent, error) {
	reader := codec.NewReader(payload)
	sessionID, err := reader.ReadString()
	if err != nil || sessionID == "" {
		return AuthenticatedEvent{}, ErrInvalidAuthenticatedPayload
	}
	userID, err := reader.ReadInt32()
	if err != nil {
		return AuthenticatedEvent{}, ErrInvalidAuthenticatedPayload
	}
	return AuthenticatedEvent{SessionID: sessionID, UserID: userID}, nil
}
