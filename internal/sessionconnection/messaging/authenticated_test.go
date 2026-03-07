package messaging

import (
	"errors"
	"testing"
)

// TestAuthenticatedEventCodec validates authenticated event payload codecs.
func TestAuthenticatedEventCodec(t *testing.T) {
	raw := EncodeAuthenticatedEvent("s1", 11)
	decoded, err := DecodeAuthenticatedEvent(raw)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if decoded.SessionID != "s1" || decoded.UserID != 11 {
		t.Fatalf("unexpected payload: %+v", decoded)
	}
}

// TestDecodeAuthenticatedEventInvalidPayload validates malformed payload handling.
func TestDecodeAuthenticatedEventInvalidPayload(t *testing.T) {
	_, err := DecodeAuthenticatedEvent([]byte{0, 0, 0})
	if !errors.Is(err, ErrInvalidAuthenticatedPayload) {
		t.Fatalf("expected invalid payload error, got %v", err)
	}
}
