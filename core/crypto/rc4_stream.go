package crypto

import (
	"crypto/rc4"
	"crypto/sha1"
	"fmt"
	"sync"
)

// StreamCipher defines per-connection symmetric stream behavior.
type StreamCipher struct {
	// inbound stores cipher state for incoming payloads.
	inbound *rc4.Cipher
	// outbound stores cipher state for outgoing payloads.
	outbound *rc4.Cipher
	// inboundMutex serializes incoming stream state mutation.
	inboundMutex sync.Mutex
	// outboundMutex serializes outgoing stream state mutation.
	outboundMutex sync.Mutex
}

// NewStreamCipher creates rc4 stream ciphers from one shared key.
func NewStreamCipher(shared []byte) (*StreamCipher, error) {
	if len(shared) == 0 {
		return nil, fmt.Errorf("shared key is required")
	}
	digest := sha1.Sum(shared)
	inbound, err := rc4.NewCipher(digest[:])
	if err != nil {
		return nil, err
	}
	outbound, err := rc4.NewCipher(digest[:])
	if err != nil {
		return nil, err
	}
	return &StreamCipher{inbound: inbound, outbound: outbound}, nil
}

// Encrypt applies outbound stream encryption to payload bytes.
func (cipher *StreamCipher) Encrypt(payload []byte) ([]byte, error) {
	if cipher == nil {
		return nil, fmt.Errorf("cipher is required")
	}
	output := make([]byte, len(payload))
	cipher.outboundMutex.Lock()
	cipher.outbound.XORKeyStream(output, payload)
	cipher.outboundMutex.Unlock()
	return output, nil
}

// Decrypt applies inbound stream decryption to payload bytes.
func (cipher *StreamCipher) Decrypt(payload []byte) ([]byte, error) {
	if cipher == nil {
		return nil, fmt.Errorf("cipher is required")
	}
	output := make([]byte, len(payload))
	cipher.inboundMutex.Lock()
	cipher.inbound.XORKeyStream(output, payload)
	cipher.inboundMutex.Unlock()
	return output, nil
}
