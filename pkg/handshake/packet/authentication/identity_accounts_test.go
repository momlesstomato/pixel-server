package authentication

import "testing"

// TestIdentityAccountsEncodeDecode verifies identity_accounts packet round-trip behavior.
func TestIdentityAccountsEncodeDecode(t *testing.T) {
	source := IdentityAccountsPacket{Accounts: []IdentityAccount{{ID: 1, Name: "Player#1"}, {ID: 2, Name: "Player#2"}}}
	encoded, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := IdentityAccountsPacket{}
	if err := decoded.Decode(encoded); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if len(decoded.Accounts) != 2 || decoded.Accounts[1].Name != "Player#2" {
		t.Fatalf("unexpected decode result: %+v", decoded)
	}
}

// TestIdentityAccountsDecodeRejectsInvalidPayload verifies decode validation behavior.
func TestIdentityAccountsDecodeRejectsInvalidPayload(t *testing.T) {
	packet := IdentityAccountsPacket{}
	if err := packet.Decode([]byte{255, 255, 255, 255}); err == nil {
		t.Fatalf("expected decode failure for negative account count")
	}
}
