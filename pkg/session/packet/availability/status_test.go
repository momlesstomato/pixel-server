package availability

import "testing"

// TestStatusPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestStatusPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := StatusPacket{IsOpen: true, OnShutdown: false, IsAuthentic: true}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := StatusPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if !decoded.IsOpen || decoded.OnShutdown || !decoded.IsAuthentic {
		t.Fatalf("unexpected decoded payload: %#v", decoded)
	}
}
