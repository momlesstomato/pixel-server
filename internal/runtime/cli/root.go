package cli

import "github.com/spf13/cobra"

// NewRootCommand creates the root CLI command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "pixelsv",
		Short:         "pixelsv core runtime",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newServeCommand())
	return cmd
}
