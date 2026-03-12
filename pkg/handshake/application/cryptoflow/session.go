package cryptoflow

import (
	"crypto/rsa"
	"fmt"

	corecrypto "github.com/momlesstomato/pixel-server/core/crypto"
)

// Session defines per-connection encryption handshake behavior.
type Session struct {
	// exchange stores generated diffie key-exchange state.
	exchange *corecrypto.DiffieHellman
	// privateKey stores generated RSA private key.
	privateKey *rsa.PrivateKey
	// serverClientEncryption stores completion response encryption flag.
	serverClientEncryption bool
	// stream stores rc4 stream once key exchange is completed.
	stream *corecrypto.StreamCipher
}

// NewSession creates one per-connection encryption handshake session.
func NewSession(options Options) (*Session, error) {
	exchange, err := corecrypto.NewDiffieHellman(options.PrimeBits, options.Random)
	if err != nil {
		return nil, err
	}
	key, err := corecrypto.GeneratePrivateKey(options.RSABits, options.Random)
	if err != nil {
		return nil, err
	}
	return &Session{exchange: exchange, privateKey: key, serverClientEncryption: options.ServerClientEncryption}, nil
}

// Begin returns init_diffie response fields for current session.
func (session *Session) Begin() (InitResponse, error) {
	if session == nil || session.exchange == nil {
		return InitResponse{}, fmt.Errorf("crypto session is required")
	}
	return InitResponse{EncryptedPrime: session.exchange.Prime().String(), EncryptedGenerator: session.exchange.Generator().String()}, nil
}

// Complete finalizes key exchange and returns complete_diffie response fields.
func (session *Session) Complete(clientPublic string) (CompleteResponse, *corecrypto.StreamCipher, error) {
	if session == nil || session.exchange == nil {
		return CompleteResponse{}, nil, fmt.Errorf("crypto session is required")
	}
	decrypted := corecrypto.DecodeClientPublicKey(session.privateKey, clientPublic)
	parsed, err := corecrypto.ParsePublicKey(decrypted)
	if err != nil {
		return CompleteResponse{}, nil, err
	}
	stream, err := corecrypto.NewStreamCipher(session.exchange.DeriveSharedKey(parsed))
	if err != nil {
		return CompleteResponse{}, nil, err
	}
	session.stream = stream
	response := CompleteResponse{EncryptedPublicKey: session.exchange.PublicKey().String(), ServerClientEncryption: session.serverClientEncryption}
	return response, stream, nil
}

// Cipher returns stream cipher when key exchange has completed.
func (session *Session) Cipher() *corecrypto.StreamCipher {
	if session == nil {
		return nil
	}
	return session.stream
}
