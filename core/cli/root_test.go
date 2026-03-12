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

// TestNewRootCommandRegistersSSO verifies sso command composition.
func TestNewRootCommandRegistersSSO(t *testing.T) {
	command := NewRootCommand(Dependencies{})
	sso, _, err := command.Find([]string{"sso"})
	if err != nil {
		t.Fatalf("expected sso command lookup success, got %v", err)
	}
	if sso == nil || sso.Name() != "sso" {
		t.Fatalf("expected sso command to be registered")
	}
}
