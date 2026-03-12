package realtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// GenerateConnectionID creates one connection identifier string.
func GenerateConnectionID(source io.Reader) (string, error) {
	reader := source
	if reader == nil {
		reader = rand.Reader
	}
	buffer := make([]byte, 16)
	if _, err := io.ReadFull(reader, buffer); err != nil {
		return "", fmt.Errorf("generate connection id: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
