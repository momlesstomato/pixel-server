package navigation

import "testing"

// TestFirstLoginOfDayPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestFirstLoginOfDayPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := FirstLoginOfDayPacket{IsFirstLogin: true}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := FirstLoginOfDayPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if !decoded.IsFirstLogin {
		t.Fatalf("expected first login flag true")
	}
}
