package command

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	"github.com/momlesstomato/pixel-server/pkg/moderation/domain"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/store"
	"github.com/spf13/cobra"
)

// openStore loads configuration and creates an action store.
func openStore(opts *options) (*store.ActionStore, func(), error) {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: opts.envFile, EnvPrefix: opts.envPrefix})
	if err != nil {
		return nil, nil, err
	}
	database, err := postgrescore.NewClient(loaded.PostgreSQL)
	if err != nil {
		return nil, nil, err
	}
	sqlDB, err := database.DB()
	if err != nil {
		return nil, nil, err
	}
	s, err := store.NewActionStore(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	return s, func() { _ = sqlDB.Close() }, nil
}

// newListCommand creates the moderation list subcommand.
func newListCommand(deps Dependencies, opts *options) *cobra.Command {
	var scope, actionType string
	var userID int
	var active bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List moderation actions",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, cleanup, err := openStore(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			filter := domain.ListFilter{Limit: 50}
			if scope != "" {
				filter.Scope = domain.ActionScope(scope)
			}
			if actionType != "" {
				filter.ActionType = domain.ActionType(actionType)
			}
			if userID > 0 {
				filter.TargetUserID = userID
			}
			if active {
				b := true
				filter.Active = &b
			}
			actions, err := s.List(context.Background(), filter)
			if err != nil {
				return err
			}
			for _, a := range actions {
				fmt.Fprintf(deps.Output, "[%d] %s/%s target=%d issuer=%d active=%t reason=%q created=%s\n",
					a.ID, a.Scope, a.ActionType, a.TargetUserID, a.IssuerID, a.Active, a.Reason, a.CreatedAt.Format(time.RFC3339))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "", "Filter by scope (room/hotel)")
	cmd.Flags().StringVar(&actionType, "type", "", "Filter by action type")
	cmd.Flags().IntVar(&userID, "user-id", 0, "Filter by target user ID")
	cmd.Flags().BoolVar(&active, "active", false, "Show only active actions")
	return cmd
}

// newBanCommand creates the moderation ban subcommand.
func newBanCommand(deps Dependencies, opts *options) *cobra.Command {
	var userID, duration int
	var reason, ip, machineID string
	cmd := &cobra.Command{
		Use:   "ban",
		Short: "Create a hotel ban",
		RunE: func(_ *cobra.Command, _ []string) error {
			if userID <= 0 {
				return fmt.Errorf("--user-id is required")
			}
			s, cleanup, err := openStore(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			action := &domain.Action{
				Scope: domain.ScopeHotel, ActionType: domain.TypeBan,
				TargetUserID: userID, IssuerID: 0, Reason: reason,
				DurationMinutes: duration, IPAddress: ip, MachineID: machineID,
				Active: true,
			}
			if duration > 0 {
				exp := time.Now().Add(time.Duration(duration) * time.Minute)
				action.ExpiresAt = &exp
			}
			if err := s.Create(context.Background(), action); err != nil {
				return err
			}
			fmt.Fprintf(deps.Output, "Created ban #%d for user %d\n", action.ID, userID)
			return nil
		},
	}
	cmd.Flags().IntVar(&userID, "user-id", 0, "Target user ID")
	cmd.Flags().IntVar(&duration, "duration", 0, "Duration in minutes (0=permanent)")
	cmd.Flags().StringVar(&reason, "reason", "", "Ban reason")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address to ban")
	cmd.Flags().StringVar(&machineID, "machine-id", "", "Machine ID to ban")
	return cmd
}

// newUnbanCommand creates the moderation unban subcommand.
func newUnbanCommand(deps Dependencies, opts *options) *cobra.Command {
	var actionID int64
	cmd := &cobra.Command{
		Use:   "unban",
		Short: "Deactivate a ban action",
		RunE: func(_ *cobra.Command, _ []string) error {
			if actionID <= 0 {
				return fmt.Errorf("--action-id is required")
			}
			s, cleanup, err := openStore(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			if err := s.Deactivate(context.Background(), actionID, 0); err != nil {
				return err
			}
			fmt.Fprintf(deps.Output, "Deactivated action #%d\n", actionID)
			return nil
		},
	}
	cmd.Flags().Int64Var(&actionID, "action-id", 0, "Action ID to deactivate")
	return cmd
}

// newHistoryCommand creates the moderation history subcommand.
func newHistoryCommand(deps Dependencies, opts *options) *cobra.Command {
	var userID int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show user moderation history",
		RunE: func(_ *cobra.Command, _ []string) error {
			if userID <= 0 {
				return fmt.Errorf("--user-id is required")
			}
			s, cleanup, err := openStore(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			actions, err := s.List(context.Background(), domain.ListFilter{TargetUserID: userID, Limit: 100})
			if err != nil {
				return err
			}
			for _, a := range actions {
				status := "active"
				if !a.Active {
					status = "inactive"
				}
				fmt.Fprintf(deps.Output, "[%d] %s %s/%s reason=%q %s\n",
					a.ID, status, a.Scope, a.ActionType, a.Reason, a.CreatedAt.Format(time.RFC3339))
			}
			if len(actions) == 0 {
				fmt.Fprintln(deps.Output, "No moderation history found.")
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&userID, "user-id", 0, "Target user ID")
	return cmd
}

// newAlertsCommand creates the moderation alerts registry subcommand.
func newAlertsCommand(deps Dependencies, opts *options) *cobra.Command {
	var scope string
	var issuerID int
	var userID int
	var roomID int
	var active bool
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "List moderation alert registry entries",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, cleanup, err := openStore(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			filter := domain.ListFilter{ActionType: domain.TypeWarn, Limit: 50}
			if scope != "" {
				filter.Scope = domain.ActionScope(scope)
			}
			if issuerID > 0 {
				filter.IssuerID = issuerID
			}
			if userID > 0 {
				filter.TargetUserID = userID
			}
			if roomID > 0 {
				filter.RoomID = roomID
			}
			if active {
				b := true
				filter.Active = &b
			}
			actions, err := s.List(context.Background(), filter)
			if err != nil {
				return err
			}
			for _, action := range actions {
				fmt.Fprintf(deps.Output, "[%d] %s target=%d room=%d issuer=%d active=%t reason=%q created=%s\n",
					action.ID, action.Scope, action.TargetUserID, action.RoomID, action.IssuerID, action.Active, action.Reason, action.CreatedAt.Format(time.RFC3339))
			}
			if len(actions) == 0 {
				fmt.Fprintln(deps.Output, "No moderation alerts found.")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "", "Filter by scope (room/hotel)")
	cmd.Flags().IntVar(&issuerID, "issuer-id", 0, "Filter by issuer user ID")
	cmd.Flags().IntVar(&userID, "user-id", 0, "Filter by target user ID")
	cmd.Flags().IntVar(&roomID, "room-id", 0, "Filter by room ID")
	cmd.Flags().BoolVar(&active, "active", false, "Show only active alerts")
	return cmd
}

// parsePositiveInt parses a string to a positive integer.
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid positive integer: %s", s)
	}
	return n, nil
}
