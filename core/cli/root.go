package cli

import (
	authcommand "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/command"
	permissioncommand "github.com/momlesstomato/pixel-server/pkg/permission/adapter/command"
	usercommand "github.com/momlesstomato/pixel-server/pkg/user/adapter/command"
	"github.com/spf13/cobra"
)

// Dependencies defines root command dependencies.
type Dependencies struct {
	// Serve defines serve command runtime dependencies.
	Serve ServeDependencies
	// Authentication defines SSO command runtime dependencies.
	Authentication authcommand.Dependencies
	// User defines user command runtime dependencies.
	User usercommand.Dependencies
	// Permission defines permission group command runtime dependencies.
	Permission permissioncommand.Dependencies
}

// NewRootCommand creates the root CLI command tree.
func NewRootCommand(dependencies Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "pixel-server",
		Short: "Pixel server runtime CLI",
	}
	command.AddCommand(NewServeCommand(dependencies.Serve))
	command.AddCommand(NewDBCommand())
	command.AddCommand(authcommand.NewSSOCommand(dependencies.Authentication))
	command.AddCommand(usercommand.NewUserCommand(dependencies.User))
	command.AddCommand(permissioncommand.NewGroupCommand(dependencies.Permission))
	return command
}
