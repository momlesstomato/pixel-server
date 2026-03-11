package authentication

import "testing"

// TestAuthenticationOKEncodeDecode verifies authentication.ok packet behavior.
func TestAuthenticationOKEncodeDecode(t *testing.T) {
	source := AuthenticationOKPacket{}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := AuthenticationOKPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
}
