package security

import (
	"errors"
	"io"
	"strings"
	"testing"
)

// failingReader defines deterministic read failures for entropy sources.
type failingReader struct{}

// Read always fails to emulate random source failures.
func (reader failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}

// TestMachineIDPolicyNormalizeAcceptsValidID verifies valid identifiers are kept.
func TestMachineIDPolicyNormalizeAcceptsValidID(t *testing.T) {
	value := strings.Repeat("A", 64)
	policy := NewMachineIDPolicy(strings.NewReader(strings.Repeat("b", 64)))
	normalized, err := policy.Normalize(value)
	if err != nil {
		t.Fatalf("expected normalization success, got %v", err)
	}
	if normalized != strings.Repeat("a", 64) {
		t.Fatalf("expected lower-case machine id, got %s", normalized)
	}
}

// TestMachineIDPolicyNormalizeRegeneratesOnInvalidInput verifies regeneration behavior.
func TestMachineIDPolicyNormalizeRegeneratesOnInvalidInput(t *testing.T) {
	policy := NewMachineIDPolicy(strings.NewReader(strings.Repeat("a", 32)))
	normalized, err := policy.Normalize("~legacy-id")
	if err != nil {
		t.Fatalf("expected regeneration success, got %v", err)
	}
	if normalized != strings.Repeat("61", 32) {
		t.Fatalf("expected deterministic generated id, got %s", normalized)
	}
}

// TestMachineIDPolicyNormalizeRegeneratesOnNonHex verifies non-hex input replacement.
func TestMachineIDPolicyNormalizeRegeneratesOnNonHex(t *testing.T) {
	policy := NewMachineIDPolicy(strings.NewReader(strings.Repeat("z", 32)))
	normalized, err := policy.Normalize("zzzz")
	if err != nil {
		t.Fatalf("expected regeneration success, got %v", err)
	}
	if len(normalized) != 64 {
		t.Fatalf("expected generated id length 64, got %d", len(normalized))
	}
}

// TestMachineIDPolicyNormalizeFailsWhenEntropyUnavailable verifies generation errors.
func TestMachineIDPolicyNormalizeFailsWhenEntropyUnavailable(t *testing.T) {
	policy := NewMachineIDPolicy(failingReader{})
	if _, err := policy.Normalize("invalid"); err == nil {
		t.Fatalf("expected generation error")
	}
}

// TestNewMachineIDPolicyUsesDefaultEntropy verifies constructor defaults.
func TestNewMachineIDPolicyUsesDefaultEntropy(t *testing.T) {
	policy := NewMachineIDPolicy(nil)
	normalized, err := policy.Normalize("bad")
	if err != nil {
		t.Fatalf("expected default entropy success, got %v", err)
	}
	if len(normalized) != 64 {
		t.Fatalf("expected generated id length 64, got %d", len(normalized))
	}
}

var _ io.Reader = failingReader{}
