package security

import "testing"

// TestSSOTicketEncodeDecode verifies sso_ticket packet round-trip behavior.
func TestSSOTicketEncodeDecode(t *testing.T) {
	timestamp := int32(123)
	source := SSOTicketPacket{Ticket: "ticket", Timestamp: &timestamp}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := SSOTicketPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.Ticket != source.Ticket || decoded.Timestamp == nil || *decoded.Timestamp != timestamp {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}
