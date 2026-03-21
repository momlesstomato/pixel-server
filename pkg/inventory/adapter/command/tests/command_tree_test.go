package tests

import (
	"testing"

	inventorycommand "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/command"
)

// TestNewInventoryCommandRegistersCoreSubcommands verifies inventory command tree composition.
func TestNewInventoryCommandRegistersCoreSubcommands(t *testing.T) {
	command := inventorycommand.NewInventoryCommand(inventorycommand.Dependencies{})
	paths := [][]string{{"credits-get"}, {"currencies-list"}, {"badges-list"}, {"effects-list"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
