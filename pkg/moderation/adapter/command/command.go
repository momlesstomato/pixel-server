package command

import (
	"io"

	"github.com/spf13/cobra"
)

// Dependencies defines required inputs for the moderation CLI tree.
type Dependencies struct {
	// Output stores the writer for command output.
	Output io.Writer
}

// options stores shared CLI bootstrap configuration.
type options struct {
	// envFile stores the path to the .env file.
	envFile string
	// envPrefix stores the environment variable prefix.
	envPrefix string
}

// NewModerationCommand creates the moderation command tree.
func NewModerationCommand(deps Dependencies) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "moderation",
		Short: "Moderation management commands",
	}
	cmd.PersistentFlags().StringVar(&opts.envFile, "env-file", ".env", "Environment file path")
	cmd.PersistentFlags().StringVar(&opts.envPrefix, "env-prefix", "MN", "Environment variable prefix")
	cmd.AddCommand(newListCommand(deps, opts))
	cmd.AddCommand(newBanCommand(deps, opts))
	cmd.AddCommand(newUnbanCommand(deps, opts))
	cmd.AddCommand(newHistoryCommand(deps, opts))
	cmd.AddCommand(newWordFilterCommand(deps, opts))
	cmd.AddCommand(newPresetCommand(deps, opts))
	return cmd
}
