package command

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// newIgnoreCommand creates the user ignore subcommand tree.
func newIgnoreCommand(deps Dependencies, options *options) *cobra.Command {
	command := &cobra.Command{Use: "ignore", Short: "Manage user ignore list"}
	command.AddCommand(newIgnoreListCommand(deps, options))
	command.AddCommand(newIgnoreAddCommand(deps, options))
	command.AddCommand(newIgnoreRemoveCommand(deps, options))
	return command
}

// newIgnoreListCommand creates the ignore list subcommand.
func newIgnoreListCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "list [id]", Short: "List ignored users", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		entries, err := service.ListIgnoredUsers(context.Background(), userID)
		if err != nil {
			return err
		}
		return printJSON(deps.Output, entries)
	}}
}

// newIgnoreAddCommand creates the ignore add subcommand.
func newIgnoreAddCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "add [id] [target-id]", Short: "Add user to ignore list", Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		targetID, err := parsePositiveID(args[1])
		if err != nil {
			return err
		}
		if err = service.AdminIgnoreUser(context.Background(), userID, targetID); err != nil {
			return err
		}
		_, err = fmt.Fprintln(deps.Output, "ok")
		return err
	}}
}

// newIgnoreRemoveCommand creates the ignore remove subcommand.
func newIgnoreRemoveCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "remove [id] [target-id]", Short: "Remove user from ignore list", Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		service, cleanup, err := openService(*options)
		if err != nil {
			return err
		}
		defer cleanup()
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		targetID, err := parsePositiveID(args[1])
		if err != nil {
			return err
		}
		if err = service.AdminUnignoreUser(context.Background(), userID, targetID); err != nil {
			return err
		}
		_, err = fmt.Fprintln(deps.Output, "ok")
		return err
	}}
}
