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

// TestServeCommandHasRoleFlag validates role flag wiring.
func TestServeCommandHasRoleFlag(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"serve", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	serveCmd, _, err := cmd.Find([]string{"serve"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	flag := serveCmd.Flags().Lookup("role")
	if flag == nil {
		t.Fatalf("expected role flag")
	}
	if flag.DefValue != "all" {
		t.Fatalf("unexpected role default: %s", flag.DefValue)
	}
}
