package tests

import (
	"testing"

	subscriptioncommand "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/command"
)

// TestNewSubscriptionCommandRegistersCoreSubcommands verifies subscription command tree composition.
func TestNewSubscriptionCommandRegistersCoreSubcommands(t *testing.T) {
	command := subscriptioncommand.NewSubscriptionCommand(subscriptioncommand.Dependencies{})
	paths := [][]string{{"status"}, {"club-offers"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
