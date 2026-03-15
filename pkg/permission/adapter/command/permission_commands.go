package command

import (
	"context"
	"strings"

	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	"github.com/spf13/cobra"
)

// newPermissionsCommand creates group permission management commands.
func newPermissionsCommand(deps Dependencies, options *options) *cobra.Command {
	command := &cobra.Command{Use: "perm", Short: "Manage group permissions"}
	command.AddCommand(newPermissionListCommand(deps, options))
	command.AddCommand(newPermissionAddCommand(deps, options))
	command.AddCommand(newPermissionRemoveCommand(deps, options))
	return command
}

// newPermissionListCommand creates the permission list subcommand.
func newPermissionListCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "list [group-id]", Short: "List permissions for one group", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, findErr := service.GetGroup(ctx, groupID)
			if findErr != nil {
				return findErr
			}
			return printJSON(deps.Output, value.Permissions)
		})
	}}
}

// newPermissionAddCommand creates the permission add subcommand.
func newPermissionAddCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "add [group-id] [permission] [permission...]", Short: "Add permissions to one group", Args: cobra.MinimumNArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		permissions := make([]string, 0, len(args)-1)
		for _, permission := range args[1:] {
			permissions = append(permissions, strings.TrimSpace(permission))
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, addErr := service.AddPermissions(ctx, groupID, permissions)
			if addErr != nil {
				return addErr
			}
			return printJSON(deps.Output, value)
		})
	}}
}

// newPermissionRemoveCommand creates the permission remove subcommand.
func newPermissionRemoveCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "remove [group-id] [permission]", Short: "Remove one permission from one group", Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		groupID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		permission := strings.TrimSpace(args[1])
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			value, removeErr := service.RemovePermission(ctx, groupID, permission)
			if removeErr != nil {
				return removeErr
			}
			return printJSON(deps.Output, value)
		})
	}}
}

// newAssignUserCommand creates the user group-assignment subcommand.
func newAssignUserCommand(deps Dependencies, options *options) *cobra.Command {
	return &cobra.Command{Use: "assign-user [user-id] [group-id] [group-id...]", Short: "Replace user group assignments", Args: cobra.MinimumNArgs(2), RunE: func(_ *cobra.Command, args []string) error {
		userID, err := parsePositiveID(args[0])
		if err != nil {
			return err
		}
		groupIDs := make([]int, 0, len(args)-1)
		for _, value := range args[1:] {
			groupID, parseErr := parsePositiveID(value)
			if parseErr != nil {
				return parseErr
			}
			groupIDs = append(groupIDs, groupID)
		}
		return runWithService(*options, func(ctx context.Context, service *permissionapplication.Service) error {
			access, assignErr := service.ReplaceUserGroups(ctx, userID, groupIDs)
			if assignErr != nil {
				return assignErr
			}
			return printJSON(deps.Output, access)
		})
	}}
}
