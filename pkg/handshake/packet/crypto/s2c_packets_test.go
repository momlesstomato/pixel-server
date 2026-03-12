package crypto

import "testing"

// TestServerInitDiffiePacketEncodeDecode verifies server parameter round-trip behavior.
func TestServerInitDiffiePacketEncodeDecode(t *testing.T) {
	source := ServerInitDiffiePacket{EncryptedPrime: "prime", EncryptedGenerator: "generator"}
	body, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ServerInitDiffiePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.EncryptedPrime != source.EncryptedPrime || decoded.EncryptedGenerator != source.EncryptedGenerator {
		t.Fatalf("unexpected decoded packet: %#v", decoded)
	}
}

// TestServerCompleteDiffiePacketEncodeDecode verifies completion packet round-trip behavior.
func TestServerCompleteDiffiePacketEncodeDecode(t *testing.T) {
	source := ServerCompleteDiffiePacket{EncryptedPublicKey: "server-public", ServerClientEncryption: true}
	body, err := source.Encode()
	if err != nil {
		t.Fatalf("expected encode success, got %v", err)
	}
	decoded := ServerCompleteDiffiePacket{}
	if err := decoded.Decode(body); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if decoded.EncryptedPublicKey != source.EncryptedPublicKey || !decoded.ServerClientEncryption {
		t.Fatalf("unexpected decoded packet: %#v", decoded)
	}
}
