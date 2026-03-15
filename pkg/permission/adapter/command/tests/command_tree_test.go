package tests

import (
	"testing"

	permissioncommand "github.com/momlesstomato/pixel-server/pkg/permission/adapter/command"
)

// TestNewGroupCommandRegistersCoreSubcommands verifies permission command tree composition.
func TestNewGroupCommandRegistersCoreSubcommands(t *testing.T) {
	command := permissioncommand.NewGroupCommand(permissioncommand.Dependencies{})
	paths := [][]string{{"list"}, {"get"}, {"create"}, {"update"}, {"delete"}, {"perm"}, {"assign-user"}}
	for _, path := range paths {
		value, _, err := command.Find(path)
		if err != nil || value == nil || value.Name() != path[0] {
			t.Fatalf("expected subcommand %s to exist, err=%v", path[0], err)
		}
	}
}
