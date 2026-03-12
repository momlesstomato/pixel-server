package cli

import (
	authcommand "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/command"
	"github.com/spf13/cobra"
)

// Dependencies defines root command dependencies.
type Dependencies struct {
	// Serve defines serve command runtime dependencies.
	Serve ServeDependencies
	// Authentication defines SSO command runtime dependencies.
	Authentication authcommand.Dependencies
}

// NewRootCommand creates the root CLI command tree.
func NewRootCommand(dependencies Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "pixel-server",
		Short: "Pixel server runtime CLI",
	}
	command.AddCommand(NewServeCommand(dependencies.Serve))
	command.AddCommand(authcommand.NewSSOCommand(dependencies.Authentication))
	return command
}
