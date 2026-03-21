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

// NewFurnitureCommand creates the furniture command tree.
func NewFurnitureCommand(dependencies Dependencies) *cobra.Command {
	var opts options
	command := &cobra.Command{Use: "furniture", Short: "Manage furniture data"}
	command.PersistentFlags().StringVar(&opts.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&opts.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newDefinitionsListCommand(dependencies, &opts))
	command.AddCommand(newDefinitionsGetCommand(dependencies, &opts))
	command.AddCommand(newItemsListCommand(dependencies, &opts))
	return command
}

// options defines command execution inputs.
type options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}
