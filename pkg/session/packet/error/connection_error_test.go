package error

import "testing"

// TestConnectionErrorPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestConnectionErrorPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := ConnectionErrorPacket{MessageID: 9999, ErrorCode: 2, Timestamp: "2026-03-12T10:20:30Z"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ConnectionErrorPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.MessageID != 9999 || decoded.ErrorCode != 2 || decoded.Timestamp != "2026-03-12T10:20:30Z" {
		t.Fatalf("unexpected decode payload %#v", decoded)
	}
}
