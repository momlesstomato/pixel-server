package cryptoflow

import "io"

const defaultPrimeBits = 128
const defaultRSABits = 1024

// Options defines runtime key-exchange settings.
type Options struct {
	// PrimeBits stores generated diffie prime size in bits.
	PrimeBits int
	// RSABits stores generated RSA private key size in bits.
	RSABits int
	// ServerClientEncryption stores completion response encryption flag.
	ServerClientEncryption bool
	// Random stores entropy source for key generation.
	Random io.Reader
}

// InitResponse defines server init_diffie packet payload fields.
type InitResponse struct {
	// EncryptedPrime stores encoded diffie prime value.
	EncryptedPrime string
	// EncryptedGenerator stores encoded diffie generator value.
	EncryptedGenerator string
}

// CompleteResponse defines server complete_diffie packet payload fields.
type CompleteResponse struct {
	// EncryptedPublicKey stores encoded server public key value.
	EncryptedPublicKey string
	// ServerClientEncryption indicates whether server-to-client encryption is enabled.
	ServerClientEncryption bool
}
