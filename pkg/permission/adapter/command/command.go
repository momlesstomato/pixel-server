package command

import (
	"io"

	"github.com/spf13/cobra"
)

// Dependencies defines command runtime overrides.
type Dependencies struct {
	// Output defines command output destination.
	Output io.Writer
}

// NewGroupCommand creates the permission group command tree.
func NewGroupCommand(dependencies Dependencies) *cobra.Command {
	var options options
	command := &cobra.Command{Use: "group", Short: "Manage permission groups"}
	command.PersistentFlags().StringVar(&options.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&options.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newListCommand(dependencies, &options))
	command.AddCommand(newGetCommand(dependencies, &options))
	command.AddCommand(newCreateCommand(dependencies, &options))
	command.AddCommand(newUpdateCommand(dependencies, &options))
	command.AddCommand(newDeleteCommand(dependencies, &options))
	command.AddCommand(newPermissionsCommand(dependencies, &options))
	command.AddCommand(newAssignUserCommand(dependencies, &options))
	return command
}

// options defines command execution inputs.
type options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}
