package cli

import "testing"

// TestNewRootCommandIncludesServe validates command wiring.
func TestNewRootCommandIncludesServe(t *testing.T) {
	cmd := NewRootCommand()
	if cmd.Name() != "pixelsv" {
		t.Fatalf("unexpected root command name: %s", cmd.Name())
	}
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "serve" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected serve subcommand")
	}
}
