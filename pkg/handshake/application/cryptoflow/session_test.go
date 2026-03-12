package cryptoflow

import (
	"math/big"
	"testing"
)

// TestSessionBeginAndComplete verifies diffie handshake and stream creation behavior.
func TestSessionBeginAndComplete(t *testing.T) {
	session, err := NewSession(Options{ServerClientEncryption: true})
	if err != nil {
		t.Fatalf("expected session creation success, got %v", err)
	}
	initResponse, err := session.Begin()
	if err != nil {
		t.Fatalf("expected begin success, got %v", err)
	}
	prime, primeOK := new(big.Int).SetString(initResponse.EncryptedPrime, 10)
	generator, generatorOK := new(big.Int).SetString(initResponse.EncryptedGenerator, 10)
	if !primeOK || !generatorOK {
		t.Fatalf("expected decimal prime and generator")
	}
	clientPrivate := big.NewInt(5)
	clientPublic := new(big.Int).Exp(generator, clientPrivate, prime)
	completeResponse, stream, err := session.Complete(clientPublic.String())
	if err != nil {
		t.Fatalf("expected complete success, got %v", err)
	}
	if completeResponse.EncryptedPublicKey == "" || !completeResponse.ServerClientEncryption {
		t.Fatalf("unexpected completion response: %#v", completeResponse)
	}
	if stream == nil || session.Cipher() == nil {
		t.Fatalf("expected non-nil stream cipher")
	}
	encrypted, err := stream.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatalf("expected encrypt success, got %v", err)
	}
	decrypted, err := stream.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("expected decrypt success, got %v", err)
	}
	if string(decrypted) != "hello" {
		t.Fatalf("expected decrypted payload hello, got %q", string(decrypted))
	}
}

// TestSessionCompleteRejectsInvalidPublicKey verifies malformed client key rejection behavior.
func TestSessionCompleteRejectsInvalidPublicKey(t *testing.T) {
	session, err := NewSession(Options{})
	if err != nil {
		t.Fatalf("expected session creation success, got %v", err)
	}
	if _, _, err := session.Complete("**invalid**"); err == nil {
		t.Fatalf("expected complete failure for invalid client key")
	}
}

// TestSessionGuards verifies nil-session guard behavior.
func TestSessionGuards(t *testing.T) {
	var session *Session
	if _, err := session.Begin(); err == nil {
		t.Fatalf("expected begin failure for nil session")
	}
	if _, _, err := session.Complete("123"); err == nil {
		t.Fatalf("expected complete failure for nil session")
	}
	if session.Cipher() != nil {
		t.Fatalf("expected nil cipher for nil session")
	}
}
