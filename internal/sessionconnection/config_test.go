package sessionconnection

import "testing"

// TestDefaultConfig validates default config values are valid.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestConfigValidateInvalid validates invalid configuration paths.
func TestConfigValidateInvalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.PingInterval = 0
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}
