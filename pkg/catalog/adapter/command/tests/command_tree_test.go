package tests

import (
	"testing"

	catalogcommand "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/command"
)

// TestNewCatalogCommandRegistersCoreSubcommands verifies catalog command tree composition.
func TestNewCatalogCommandRegistersCoreSubcommands(t *testing.T) {
	command := catalogcommand.NewCatalogCommand(catalogcommand.Dependencies{})
	paths := [][]string{{"pages-list"}, {"pages-get"}, {"offers-list"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
