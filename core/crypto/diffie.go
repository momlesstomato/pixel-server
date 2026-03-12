package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"strings"
)

const defaultPrimeBits = 128

// DiffieHellman stores per-connection diffie parameters and keys.
type DiffieHellman struct {
	// prime stores generated diffie modulus.
	prime *big.Int
	// generator stores generated diffie base value.
	generator *big.Int
	// privateKey stores generated private exponent.
	privateKey *big.Int
	// publicKey stores generated public value.
	publicKey *big.Int
}

// NewDiffieHellman creates one diffie key-exchange context.
func NewDiffieHellman(bits int, source io.Reader) (*DiffieHellman, error) {
	reader := source
	if reader == nil {
		reader = rand.Reader
	}
	size := bits
	if size <= 0 {
		size = defaultPrimeBits
	}
	prime, err := rand.Prime(reader, size)
	if err != nil {
		return nil, err
	}
	generator := big.NewInt(2)
	privateKey, err := rand.Int(reader, prime)
	if err != nil {
		return nil, err
	}
	publicKey := new(big.Int).Exp(generator, privateKey, prime)
	return &DiffieHellman{prime: prime, generator: generator, privateKey: privateKey, publicKey: publicKey}, nil
}

// Prime returns generated prime value.
func (exchange *DiffieHellman) Prime() *big.Int {
	if exchange == nil || exchange.prime == nil {
		return nil
	}
	return new(big.Int).Set(exchange.prime)
}

// Generator returns generated generator value.
func (exchange *DiffieHellman) Generator() *big.Int {
	if exchange == nil || exchange.generator == nil {
		return nil
	}
	return new(big.Int).Set(exchange.generator)
}

// PublicKey returns generated server public key value.
func (exchange *DiffieHellman) PublicKey() *big.Int {
	if exchange == nil || exchange.publicKey == nil {
		return nil
	}
	return new(big.Int).Set(exchange.publicKey)
}

// DeriveSharedKey creates shared secret bytes from one client public key.
func (exchange *DiffieHellman) DeriveSharedKey(clientPublic *big.Int) []byte {
	if exchange == nil || exchange.privateKey == nil || exchange.prime == nil || clientPublic == nil {
		return nil
	}
	shared := new(big.Int).Exp(clientPublic, exchange.privateKey, exchange.prime)
	value := shared.Bytes()
	if len(value) == 0 {
		return []byte{0}
	}
	return value
}

// ParsePublicKey parses decimal, hex, or base64-encoded public key strings.
func ParsePublicKey(value string) (*big.Int, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, fmt.Errorf("public key is required")
	}
	if strings.HasPrefix(trimmed, "0x") || strings.HasPrefix(trimmed, "0X") {
		if parsed, ok := new(big.Int).SetString(trimmed[2:], 16); ok {
			return parsed, nil
		}
	}
	if parsed, ok := new(big.Int).SetString(trimmed, 10); ok {
		return parsed, nil
	}
	if parsed, ok := new(big.Int).SetString(trimmed, 16); ok {
		return parsed, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil && len(decoded) > 0 {
		return new(big.Int).SetBytes(decoded), nil
	}
	return nil, fmt.Errorf("public key format is invalid")
}
