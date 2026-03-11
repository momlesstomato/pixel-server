package session

import "testing"

// TestClientPingEncodeDecode verifies client.ping packet behavior.
func TestClientPingEncodeDecode(t *testing.T) {
	source := ClientPingPacket{}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientPingPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}

// TestClientPingDecodeRejectsNonEmptyBody verifies empty-body validation behavior.
func TestClientPingDecodeRejectsNonEmptyBody(t *testing.T) {
	packet := ClientPingPacket{}
	if err := packet.Decode([]byte{1}); err == nil {
		t.Fatalf("expected decode failure for non-empty body")
	}
}
