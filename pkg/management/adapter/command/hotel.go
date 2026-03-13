package command

import (
	"context"
	"fmt"
	"io"
	"os"

	statusdomain "github.com/momlesstomato/pixel-server/pkg/status/domain"
	"github.com/spf13/cobra"
)

// HotelDependencies defines hotel CLI command dependencies.
type HotelDependencies struct {
	// Manager exposes hotel status operations for the CLI.
	Manager HotelManager
	// Output defines writer for CLI output.
	Output io.Writer
}

// HotelManager defines hotel status management behavior for CLI commands.
type HotelManager interface {
	// Current returns active hotel status snapshot.
	Current(context.Context) (statusdomain.HotelStatus, error)
	// ScheduleClose transitions hotel into closing state.
	ScheduleClose(context.Context, int32, int32, bool) (statusdomain.HotelStatus, error)
	// Reopen transitions hotel into open state.
	Reopen(context.Context) (statusdomain.HotelStatus, error)
}

// NewHotelCommand creates the hotel management subcommand tree.
func NewHotelCommand(deps HotelDependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "hotel",
		Short: "Manage hotel status",
	}
	command.AddCommand(newHotelStatusCommand(deps))
	command.AddCommand(newHotelCloseCommand(deps))
	command.AddCommand(newHotelReopenCommand(deps))
	return command
}

// newHotelStatusCommand creates the hotel status subcommand.
func newHotelStatusCommand(deps HotelDependencies) *cobra.Command {
	return &cobra.Command{
		Use: "status", Short: "Show current hotel status",
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeHotelStatus(deps)
		},
	}
}

// newHotelCloseCommand creates the hotel close subcommand.
func newHotelCloseCommand(deps HotelDependencies) *cobra.Command {
	var minutes, duration int32
	var throwUsers bool
	command := &cobra.Command{
		Use: "close", Short: "Schedule hotel closing",
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeHotelClose(deps, minutes, duration, throwUsers)
		},
	}
	command.Flags().Int32Var(&minutes, "minutes", 5, "Minutes until close")
	command.Flags().Int32Var(&duration, "duration", 15, "Maintenance duration in minutes")
	command.Flags().BoolVar(&throwUsers, "throw-users", false, "Disconnect users at close")
	return command
}

// newHotelReopenCommand creates the hotel reopen subcommand.
func newHotelReopenCommand(deps HotelDependencies) *cobra.Command {
	return &cobra.Command{
		Use: "reopen", Short: "Reopen hotel immediately",
		RunE: func(_ *cobra.Command, _ []string) error {
			return executeHotelReopen(deps)
		},
	}
}

// executeHotelStatus prints current hotel status.
func executeHotelStatus(deps HotelDependencies) error {
	out := deps.Output
	if out == nil {
		out = os.Stdout
	}
	status, err := deps.Manager.Current(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "state=%s\n", status.State)
	if status.CloseAt != nil {
		fmt.Fprintf(out, "close_at=%s\n", status.CloseAt.UTC())
	}
	if status.ReopenAt != nil {
		fmt.Fprintf(out, "reopen_at=%s\n", status.ReopenAt.UTC())
	}
	return nil
}

// executeHotelClose schedules hotel close.
func executeHotelClose(deps HotelDependencies, minutes int32, duration int32, throwUsers bool) error {
	out := deps.Output
	if out == nil {
		out = os.Stdout
	}
	status, err := deps.Manager.ScheduleClose(context.Background(), minutes, duration, throwUsers)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "scheduled: state=%s\n", status.State)
	return nil
}

// executeHotelReopen reopens the hotel.
func executeHotelReopen(deps HotelDependencies) error {
	out := deps.Output
	if out == nil {
		out = os.Stdout
	}
	status, err := deps.Manager.Reopen(context.Background())
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "reopened: state=%s\n", status.State)
	return nil
}
