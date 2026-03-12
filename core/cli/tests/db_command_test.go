package tests

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/cli"
)

// TestNewDBCommandRegistersSubcommands verifies migration and seed command composition.
func TestNewDBCommandRegistersSubcommands(t *testing.T) {
	command := cli.NewDBCommand()
	expected := []string{"migrate-up", "migrate-down", "seed-up", "seed-down"}
	for _, name := range expected {
		child, _, err := command.Find([]string{name})
		if err != nil {
			t.Fatalf("expected command %q lookup success, got %v", name, err)
		}
		if child == nil || child.Name() != name {
			t.Fatalf("expected command %q to be registered", name)
		}
	}
}
