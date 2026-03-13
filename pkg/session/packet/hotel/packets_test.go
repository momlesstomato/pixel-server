package hotel

import "testing"

// TestWillClosePacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestWillClosePacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := WillClosePacket{Minutes: 5}
	body, _ := packet.Encode()
	decoded := WillClosePacket{}
	if err := decoded.Decode(body); err != nil || decoded.Minutes != 5 {
		t.Fatalf("unexpected decode result: %#v err=%v", decoded, err)
	}
}

// TestMaintenancePacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestMaintenancePacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := MaintenancePacket{IsInMaintenance: true, MinutesUntilChange: 4, Duration: 15}
	body, _ := packet.Encode()
	decoded := MaintenancePacket{}
	if err := decoded.Decode(body); err != nil || !decoded.IsInMaintenance || decoded.MinutesUntilChange != 4 || decoded.Duration != 15 {
		t.Fatalf("unexpected decode result: %#v err=%v", decoded, err)
	}
}

// TestClosesAndOpensAtPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestClosesAndOpensAtPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := ClosesAndOpensAtPacket{OpenHour: 7, OpenMinute: 30, UserThrownOutAtClose: true}
	body, _ := packet.Encode()
	decoded := ClosesAndOpensAtPacket{}
	if err := decoded.Decode(body); err != nil || decoded.OpenHour != 7 || decoded.OpenMinute != 30 || !decoded.UserThrownOutAtClose {
		t.Fatalf("unexpected decode result: %#v err=%v", decoded, err)
	}
}

// TestClosedAndOpensPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestClosedAndOpensPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := ClosedAndOpensPacket{OpenHour: 9, OpenMinute: 15}
	body, _ := packet.Encode()
	decoded := ClosedAndOpensPacket{}
	if err := decoded.Decode(body); err != nil || decoded.OpenHour != 9 || decoded.OpenMinute != 15 {
		t.Fatalf("unexpected decode result: %#v err=%v", decoded, err)
	}
}
