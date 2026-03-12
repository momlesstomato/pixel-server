package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"testing"
)

// TestDiffieHellmanExchange verifies shared-key derivation behavior.
func TestDiffieHellmanExchange(t *testing.T) {
	exchange, err := NewDiffieHellman(64, rand.Reader)
	if err != nil {
		t.Fatalf("expected exchange creation success, got %v", err)
	}
	prime := exchange.Prime()
	generator := exchange.Generator()
	if prime == nil || generator == nil || exchange.PublicKey() == nil {
		t.Fatalf("expected generated parameters")
	}
	clientPrivate := big.NewInt(5)
	clientPublic := new(big.Int).Exp(generator, clientPrivate, prime)
	shared := exchange.DeriveSharedKey(clientPublic)
	if len(shared) == 0 {
		t.Fatalf("expected shared key bytes")
	}
	if len((&DiffieHellman{}).DeriveSharedKey(clientPublic)) != 0 {
		t.Fatalf("expected nil guard shared key output")
	}
}

// TestParsePublicKeyCoversFormats verifies supported key string formats.
func TestParsePublicKeyCoversFormats(t *testing.T) {
	cases := []string{"12345", "3039", "0x3039", base64.StdEncoding.EncodeToString([]byte{0x30, 0x39})}
	for _, sample := range cases {
		if _, err := ParsePublicKey(sample); err != nil {
			t.Fatalf("expected parse success for %q: %v", sample, err)
		}
	}
	if _, err := ParsePublicKey(" "); err == nil {
		t.Fatalf("expected parse failure for empty input")
	}
	if _, err := ParsePublicKey("***"); err == nil {
		t.Fatalf("expected parse failure for malformed input")
	}
}

// TestRSAHelpers verifies private key generation and decode fallback behavior.
func TestRSAHelpers(t *testing.T) {
	privateKey, err := GeneratePrivateKey(1024, rand.Reader)
	if err != nil {
		t.Fatalf("expected private key generation success, got %v", err)
	}
	payload, err := rsa.EncryptPKCS1v15(rand.Reader, &privateKey.PublicKey, []byte("123"))
	if err != nil {
		t.Fatalf("expected rsa encrypt success, got %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(payload)
	if decoded := DecodeClientPublicKey(privateKey, encoded); decoded != "123" {
		t.Fatalf("expected decoded rsa payload 123, got %q", decoded)
	}
	if fallback := DecodeClientPublicKey(privateKey, "invalid"); fallback != "invalid" {
		t.Fatalf("expected fallback payload unchanged, got %q", fallback)
	}
}

// TestStreamCipherGuards verifies stream constructor and nil-receiver guard behavior.
func TestStreamCipherGuards(t *testing.T) {
	if _, err := NewStreamCipher(nil); err == nil {
		t.Fatalf("expected shared key precondition error")
	}
	cipher, err := NewStreamCipher([]byte("secret"))
	if err != nil {
		t.Fatalf("expected stream creation success, got %v", err)
	}
	encrypted, err := cipher.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatalf("expected encrypt success, got %v", err)
	}
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil || string(decrypted) != "hello" {
		t.Fatalf("expected decrypt success, got %q err=%v", string(decrypted), err)
	}
	var missing *StreamCipher
	if _, err := missing.Encrypt([]byte("x")); err == nil {
		t.Fatalf("expected nil cipher encrypt failure")
	}
	if _, err := missing.Decrypt([]byte("x")); err == nil {
		t.Fatalf("expected nil cipher decrypt failure")
	}
}
