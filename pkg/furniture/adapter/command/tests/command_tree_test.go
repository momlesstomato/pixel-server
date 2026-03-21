package tests

import (
	"testing"

	furniturecommand "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/command"
)

// TestNewFurnitureCommandRegistersCoreSubcommands verifies furniture command tree composition.
func TestNewFurnitureCommandRegistersCoreSubcommands(t *testing.T) {
	command := furniturecommand.NewFurnitureCommand(furniturecommand.Dependencies{})
	paths := [][]string{{"definitions-list"}, {"definitions-get"}, {"items-list"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
