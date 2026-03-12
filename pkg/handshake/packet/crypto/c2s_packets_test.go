package crypto

import "testing"

// TestClientInitDiffiePacketEncodeDecode verifies empty init packet behavior.
func TestClientInitDiffiePacketEncodeDecode(t *testing.T) {
	packet := ClientInitDiffiePacket{}
	body, err := packet.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientInitDiffiePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if err := decoded.Decode([]byte{1}); err == nil {
		t.Fatalf("expected decode failure for non-empty body")
	}
}

// TestClientCompleteDiffiePacketEncodeDecode verifies client public key round-trip behavior.
func TestClientCompleteDiffiePacketEncodeDecode(t *testing.T) {
	source := ClientCompleteDiffiePacket{EncryptedPublicKey: "client-public"}
	body, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ClientCompleteDiffiePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.EncryptedPublicKey != source.EncryptedPublicKey {
		t.Fatalf("expected public key %q, got %q", source.EncryptedPublicKey, decoded.EncryptedPublicKey)
	}
}
