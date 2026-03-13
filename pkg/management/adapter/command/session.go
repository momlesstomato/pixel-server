package command

import (
	"fmt"
	"io"
	"os"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/spf13/cobra"
)

// SessionDependencies defines session CLI command dependencies.
type SessionDependencies struct {
	// Registry exposes session listing behavior for the CLI.
	Registry coreconnection.SessionRegistry
	// Output defines writer for CLI output.
	Output io.Writer
}

// NewSessionCommand creates the session management subcommand tree.
func NewSessionCommand(deps SessionDependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "session",
		Short: "Manage active sessions",
	}
	command.AddCommand(newSessionListCommand(deps))
	command.AddCommand(newSessionKickCommand(deps))
	return command
}

// newSessionListCommand creates the session list subcommand.
func newSessionListCommand(deps SessionDependencies) *cobra.Command {
	var instance string
	command := &cobra.Command{
		Use:   "list",
		Short: "List all active sessions",
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeSessionList(deps, instance)
		},
	}
	command.Flags().StringVar(&instance, "instance", "", "Filter sessions by instance ID")
	return command
}

// newSessionKickCommand creates the session kick subcommand.
func newSessionKickCommand(deps SessionDependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "kick [connID]",
		Short: "Disconnect one session by connection ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return executeSessionKick(deps, args[0])
		},
	}
}

// executeSessionList lists sessions optionally filtered by instance.
func executeSessionList(deps SessionDependencies, instance string) error {
	out := deps.Output
	if out == nil {
		out = os.Stdout
	}
	sessions, err := deps.Registry.ListAll()
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if instance != "" && s.InstanceID != instance {
			continue
		}
		fmt.Fprintf(out, "conn=%s user=%d instance=%s state=%d\n", s.ConnID, s.UserID, s.InstanceID, s.State)
	}
	fmt.Fprintf(out, "total: %d\n", len(sessions))
	return nil
}

// executeSessionKick disconnects one session by removing it from the registry.
func executeSessionKick(deps SessionDependencies, connID string) error {
	out := deps.Output
	if out == nil {
		out = os.Stdout
	}
	_, found := deps.Registry.FindByConnID(connID)
	if !found {
		return fmt.Errorf("session not found: %s", connID)
	}
	deps.Registry.Remove(connID)
	fmt.Fprintf(out, "disconnected: %s\n", connID)
	return nil
}
