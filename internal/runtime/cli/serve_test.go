package cli

import (
	"strings"
	"testing"
)

// TestServeCommandRequiresAPIKey validates API key requirement propagation.
func TestServeCommandRequiresAPIKey(t *testing.T) {
	t.Setenv("API_KEY", "")
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"serve", "--env-file", "testdata/missing.env"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "api key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
