package notification

import "testing"

// TestGenericErrorPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestGenericErrorPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := GenericErrorPacket{ErrorCode: -400}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := GenericErrorPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.ErrorCode != -400 {
		t.Fatalf("unexpected error code %d", decoded.ErrorCode)
	}
}

// TestGenericAlertPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestGenericAlertPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := GenericAlertPacket{Message: "hello"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := GenericAlertPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Message != "hello" {
		t.Fatalf("unexpected message %q", decoded.Message)
	}
}

// TestModerationCautionPacketEncodeDecodeRoundTrip verifies packet serialization behavior.
func TestModerationCautionPacketEncodeDecodeRoundTrip(t *testing.T) {
	packet := ModerationCautionPacket{Message: "warn", Detail: "details"}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ModerationCautionPacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Message != "warn" || decoded.Detail != "details" {
		t.Fatalf("unexpected caution payload %#v", decoded)
	}
}
