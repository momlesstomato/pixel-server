package authentication

import "testing"

// TestDisconnectReasonPacketEncodeDecode verifies disconnect_reason packet behavior.
func TestDisconnectReasonPacketEncodeDecode(t *testing.T) {
	encoded, err := DisconnectReasonPacket{Reason: DisconnectReasonInvalidLoginTicket}.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := DisconnectReasonPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Reason != DisconnectReasonInvalidLoginTicket {
		t.Fatalf("expected reason %d, got %d", DisconnectReasonInvalidLoginTicket, decoded.Reason)
	}
}

// TestDisconnectReasonPacketRejectsTrailingBytes verifies decode validation behavior.
func TestDisconnectReasonPacketRejectsTrailingBytes(t *testing.T) {
	packet := DisconnectReasonPacket{}
	if err := packet.Decode([]byte{0, 0, 0, 22, 0}); err == nil {
		t.Fatalf("expected decode failure for trailing bytes")
	}
}
