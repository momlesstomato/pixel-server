package navigation

import "testing"

// TestDesktopViewRequestPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestDesktopViewRequestPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := DesktopViewRequestPacket{}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := DesktopViewRequestPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}
