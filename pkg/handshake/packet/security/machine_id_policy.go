package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var machineIDPattern = regexp.MustCompile("^[0-9a-fA-F]{64}$")

// MachineIDPolicy defines machine identifier normalization behavior.
type MachineIDPolicy struct {
	// random provides entropy source for generated identifiers.
	random io.Reader
}

// NewMachineIDPolicy creates machine identifier policy behavior.
func NewMachineIDPolicy(random io.Reader) *MachineIDPolicy {
	source := random
	if source == nil {
		source = rand.Reader
	}
	return &MachineIDPolicy{random: source}
}

// Normalize validates machine id and regenerates one when input is invalid.
func (policy *MachineIDPolicy) Normalize(machineID string) (string, error) {
	trimmed := strings.TrimSpace(machineID)
	if trimmed != "" && !strings.HasPrefix(trimmed, "~") && machineIDPattern.MatchString(trimmed) {
		return strings.ToLower(trimmed), nil
	}
	buffer := make([]byte, 32)
	if _, err := io.ReadFull(policy.random, buffer); err != nil {
		return "", fmt.Errorf("generate machine id: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
