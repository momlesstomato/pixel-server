package session

import "testing"

// TestClientDisconnectEncodeDecode verifies client.disconnect packet behavior.
func TestClientDisconnectEncodeDecode(t *testing.T) {
	source := ClientDisconnectPacket{}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientDisconnectPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}

// TestClientDisconnectDecodeRejectsNonEmptyBody verifies empty-body validation behavior.
func TestClientDisconnectDecodeRejectsNonEmptyBody(t *testing.T) {
	packet := ClientDisconnectPacket{}
	if err := packet.Decode([]byte{1}); err == nil {
		t.Fatalf("expected decode failure for non-empty body")
	}
}
