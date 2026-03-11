package session

import "testing"

// TestClientPongEncodeDecode verifies client.pong packet behavior.
func TestClientPongEncodeDecode(t *testing.T) {
	source := ClientPongPacket{}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientPongPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}

// TestClientPongDecodeRejectsNonEmptyBody verifies empty-body validation behavior.
func TestClientPongDecodeRejectsNonEmptyBody(t *testing.T) {
	packet := ClientPongPacket{}
	if err := packet.Decode([]byte{1}); err == nil {
		t.Fatalf("expected decode failure for non-empty body")
	}
}
