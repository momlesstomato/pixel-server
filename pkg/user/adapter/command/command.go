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

// NewUserCommand creates the user command tree.
func NewUserCommand(dependencies Dependencies) *cobra.Command {
	var options options
	command := &cobra.Command{Use: "user", Short: "Manage user profile data"}
	command.PersistentFlags().StringVar(&options.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&options.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newGetCommand(dependencies, &options))
	command.AddCommand(newUpdateCommand(dependencies, &options))
	command.AddCommand(newRespectCommand(dependencies, &options))
	return command
}

// options defines command execution inputs.
type options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}
