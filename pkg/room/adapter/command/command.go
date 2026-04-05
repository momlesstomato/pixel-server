package command

import (
	"io"

	"github.com/spf13/cobra"
)

// Dependencies defines room CLI command runtime dependencies.
type Dependencies struct {
	// Output overrides output stream for testing.
	Output io.Writer
}

// options holds shared CLI option flags for room subcommands.
type options struct {
	// EnvFile stores the path to the environment file.
	EnvFile string
	// EnvPrefix stores the optional environment variable prefix.
	EnvPrefix string
}

// NewRoomCommand creates the room CLI command tree.
func NewRoomCommand(deps Dependencies) *cobra.Command {
	var opts options
	command := &cobra.Command{Use: "room", Short: "Manage room data"}
	command.PersistentFlags().StringVar(&opts.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&opts.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newChatExportCommand(deps, &opts))
	return command
}