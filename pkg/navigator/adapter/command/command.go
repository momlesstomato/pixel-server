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

// NewNavigatorCommand creates the navigator command tree.
func NewNavigatorCommand(dependencies Dependencies) *cobra.Command {
	var opts options
	command := &cobra.Command{Use: "navigator", Short: "Manage navigator data"}
	command.PersistentFlags().StringVar(&opts.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&opts.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newCategoriesListCommand(dependencies, &opts))
	command.AddCommand(newRoomGetCommand(dependencies, &opts))
	return command
}

// options defines command execution inputs.
type options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}
