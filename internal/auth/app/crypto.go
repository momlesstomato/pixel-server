package app

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

const groupPrimeHex = "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A63A3620FFFFFFFFFFFFFFFF"

// handshakeCrypto stores cryptographic helpers for diffie exchange.
type handshakeCrypto struct {
	// prime is the static diffie prime modulus.
	prime *big.Int
	// generator is the static diffie generator.
	generator *big.Int
	// signer signs outbound diffie fields for client verification.
	signer *rsa.PrivateKey
}

// diffieSession stores per-session diffie key data.
type diffieSession struct {
	// PrivateKey is the server private exponent.
	PrivateKey *big.Int
	// PublicKey is the server public key.
	PublicKey *big.Int
	// SharedKey is computed after complete_diffie.
	SharedKey *big.Int
}

// newHandshakeCrypto builds default handshake cryptographic primitives.
func newHandshakeCrypto() *handshakeCrypto {
	prime, _ := new(big.Int).SetString(groupPrimeHex, 16)
	signer, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		signer = nil
	}
	return &handshakeCrypto{prime: prime, generator: big.NewInt(2), signer: signer}
}

// startDiffie creates a new diffie session and signed init payload values.
func (c *handshakeCrypto) startDiffie() (*diffieSession, string, string, error) {
	if c == nil || c.prime == nil || c.generator == nil {
		return nil, "", "", fmt.Errorf("handshake crypto is not initialized")
	}
	limit := new(big.Int).Sub(c.prime, big.NewInt(3))
	private, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return nil, "", "", err
	}
	private.Add(private, big.NewInt(2))
	public := new(big.Int).Exp(c.generator, private, c.prime)
	session := &diffieSession{PrivateKey: private, PublicKey: public}
	return session, c.sign(c.prime.String()), c.sign(c.generator.String()), nil
}

// completeDiffie consumes one client public key and computes shared secret.
func (c *handshakeCrypto) completeDiffie(state *diffieSession, encryptedPublicKey string) (string, error) {
	if state == nil || state.PrivateKey == nil {
		return "", ErrDiffieNotInitialized
	}
	clientPublic, err := c.decodeClientPublicKey(encryptedPublicKey)
	if err != nil {
		return "", err
	}
	state.SharedKey = new(big.Int).Exp(clientPublic, state.PrivateKey, c.prime)
	return c.sign(state.PublicKey.String()), nil
}

// decodeClientPublicKey decodes a decimal value or RSA-encrypted decimal payload.
func (c *handshakeCrypto) decodeClientPublicKey(value string) (*big.Int, error) {
	if decimal, ok := new(big.Int).SetString(value, 10); ok {
		return decimal, nil
	}
	if c.signer == nil {
		return nil, fmt.Errorf("invalid client public key")
	}
	cipherText, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}
	plain, err := rsa.DecryptPKCS1v15(rand.Reader, c.signer, cipherText)
	if err != nil {
		return nil, err
	}
	decoded, ok := new(big.Int).SetString(string(plain), 10)
	if !ok {
		return nil, fmt.Errorf("invalid decoded client public key")
	}
	return decoded, nil
}

// sign signs one value using RSA PKCS1v15 and returns a hex signature.
func (c *handshakeCrypto) sign(value string) string {
	if c == nil || c.signer == nil {
		return value
	}
	sum := sha256.Sum256([]byte(value))
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.signer, crypto.SHA256, sum[:])
	if err != nil {
		return value
	}
	return hex.EncodeToString(signature)
}
