package navigation

import "testing"

// TestDesktopViewResponsePacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestDesktopViewResponsePacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := DesktopViewResponsePacket{}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := DesktopViewResponsePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}
