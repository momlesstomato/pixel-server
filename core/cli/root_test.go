package cli

import "testing"

// TestNewRootCommandRegistersServe verifies root command composition.
func TestNewRootCommandRegistersServe(t *testing.T) {
	command := NewRootCommand(Dependencies{})
	serve, _, err := command.Find([]string{"serve"})
	if err != nil {
		t.Fatalf("expected serve command lookup success, got %v", err)
	}
	if serve == nil || serve.Name() != "serve" {
		t.Fatalf("expected serve command to be registered")
	}
}
