package cli

import "testing"

// TestRunDBActionRejectsUnsupportedAction verifies unknown action handling.
func TestRunDBActionRejectsUnsupportedAction(t *testing.T) {
	err := runDBAction(nil, "unknown")
	if err == nil {
		t.Fatalf("expected unsupported action error")
	}
}

// TestExecuteDBActionReturnsErrorOnMissingConfig verifies config load failure handling.
func TestExecuteDBActionReturnsErrorOnMissingConfig(t *testing.T) {
	err := executeDBAction(DBOptions{EnvFile: "./does-not-exist/.env"}, "seed-up", nil, nil)
	if err == nil {
		t.Fatalf("expected configuration error")
	}
}
