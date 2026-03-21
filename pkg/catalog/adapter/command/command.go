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

// NewCatalogCommand creates the catalog command tree.
func NewCatalogCommand(dependencies Dependencies) *cobra.Command {
	var opts options
	command := &cobra.Command{Use: "catalog", Short: "Manage catalog data"}
	command.PersistentFlags().StringVar(&opts.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&opts.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newPagesListCommand(dependencies, &opts))
	command.AddCommand(newPagesGetCommand(dependencies, &opts))
	command.AddCommand(newOffersListCommand(dependencies, &opts))
	return command
}

// options defines command execution inputs.
type options struct {
	// EnvFile defines configuration file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}
