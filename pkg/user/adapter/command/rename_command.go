package command

import (
	"context"

	"github.com/spf13/cobra"
)

// newRenameCommand creates the user rename subcommand.
func newRenameCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "rename [id] [name]", Short: "Force user name change", Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		result, err := service.ForceChangeName(context.Background(), userID, args[1])
		if err != nil {
			return err
		}
		return printJSON(deps.Output, result)
	}}
}
