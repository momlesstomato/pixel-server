package command

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// newRespectCommand creates the user respect subcommand.
func newRespectCommand(deps Dependencies, options *options) *cobra.Command {
	var actorUserID int
	command := &cobra.Command{Use: "respect [target-id]", Short: "Send one user respect", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		targetUserID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		result, err := service.RecordUserRespect(context.Background(), actorUserID, targetUserID, time.Now().UTC())
		if err != nil {
			return err
		}
		return printJSON(deps.Output, result)
	}}
	command.Flags().IntVar(&actorUserID, "actor-user-id", 0, "Actor user id")
	_ = command.MarkFlagRequired("actor-user-id")
	return command
}
