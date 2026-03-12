package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
)

const defaultRSABits = 1024

// GeneratePrivateKey creates one RSA private key.
func GeneratePrivateKey(bits int, source io.Reader) (*rsa.PrivateKey, error) {
	reader := source
	if reader == nil {
		reader = rand.Reader
	}
	size := bits
	if size <= 0 {
		size = defaultRSABits
	}
	return rsa.GenerateKey(reader, size)
}

// DecodeClientPublicKey decrypts rsa-wrapped key payload and otherwise returns original value.
func DecodeClientPublicKey(privateKey *rsa.PrivateKey, encrypted string) string {
	if privateKey == nil {
		return encrypted
	}
	payload, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return encrypted
	}
	plaintext, err := rsa.DecryptPKCS1v15(nil, privateKey, payload)
	if err != nil || len(plaintext) == 0 {
		return encrypted
	}
	return string(plaintext)
}
