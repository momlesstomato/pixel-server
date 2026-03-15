package command

import (
	"context"

	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
	"github.com/spf13/cobra"
)

// newListCommand creates the group list subcommand.
func newListCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List all permission groups", RunE: func(_ *cobra.Command, _ []string) error {
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, err := service.ListGroups(ctx)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, value)
		})
	}}
}

// newGetCommand creates the group get subcommand.
func newGetCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "get [id]", Short: "Get group details by ID", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, findErr := service.GetGroup(ctx, groupID)
			if findErr != nil {
				return findErr
			}
			return printJSON(deps.Output, value)
		})
	}}
}

// newCreateCommand creates the group create subcommand.
func newCreateCommand(deps Dependencies, options *options) *cobra.Command {
	var payload permissionapplication.CreateGroupInput
	command := &cobra.Command{Use: "create [name]", Short: "Create one permission group", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		payload.Name = args[0]
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, err := service.CreateGroup(ctx, payload)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, value)
		})
	}}
	command.Flags().StringVar(&payload.DisplayName, "display", "", "Display name")
	command.Flags().IntVar(&payload.Priority, "priority", 0, "Priority")
	command.Flags().IntVar(&payload.ClubLevel, "club", 0, "Club level")
	command.Flags().IntVar(&payload.SecurityLevel, "security", 0, "Security level")
	command.Flags().BoolVar(&payload.IsAmbassador, "ambassador", false, "Ambassador role")
	command.Flags().BoolVar(&payload.IsDefault, "default", false, "Default group")
	return command
}

// newUpdateCommand creates the group update subcommand.
func newUpdateCommand(deps Dependencies, options *options) *cobra.Command {
	var display string
	var priority, club, security int
	var ambassador, isDefault bool
	command := &cobra.Command{Use: "update [id]", Short: "Update one permission group", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		patch := permissiondomain.GroupPatch{}
		if command.Flags().Changed("display") {
			patch.DisplayName = &display
		}
		if command.Flags().Changed("priority") {
			patch.Priority = &priority
		}
		if command.Flags().Changed("club") {
			patch.ClubLevel = &club
		}
		if command.Flags().Changed("security") {
			patch.SecurityLevel = &security
		}
		if command.Flags().Changed("ambassador") {
			patch.IsAmbassador = &ambassador
		}
		if command.Flags().Changed("default") {
			patch.IsDefault = &isDefault
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, updateErr := service.UpdateGroup(ctx, groupID, patch)
			if updateErr != nil {
				return updateErr
			}
			return printJSON(deps.Output, value)
		})
	}}
	command.Flags().StringVar(&display, "display", "", "Display name")
	command.Flags().IntVar(&priority, "priority", 0, "Priority")
	command.Flags().IntVar(&club, "club", 0, "Club level")
	command.Flags().IntVar(&security, "security", 0, "Security level")
	command.Flags().BoolVar(&ambassador, "ambassador", false, "Ambassador role")
	command.Flags().BoolVar(&isDefault, "default", false, "Default group")
	return command
}

// newDeleteCommand creates the group delete subcommand.
func newDeleteCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "delete [id]", Short: "Delete one permission group", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			if deleteErr := service.DeleteGroup(ctx, groupID); deleteErr != nil {
				return deleteErr
			}
			return printJSON(deps.Output, map[string]any{"deleted": groupID})
		})
	}}
}
