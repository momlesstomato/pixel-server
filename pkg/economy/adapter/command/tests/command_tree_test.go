package tests

import (
	"testing"

	economycommand "github.com/momlesstomato/pixel-server/pkg/economy/adapter/command"
)

// TestNewEconomyCommandRegistersCoreSubcommands verifies economy command tree composition.
func TestNewEconomyCommandRegistersCoreSubcommands(t *testing.T) {
	command := economycommand.NewEconomyCommand(economycommand.Dependencies{})
	paths := [][]string{{"offers-get"}, {"history-get"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
