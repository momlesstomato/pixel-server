package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	messengerstore "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/store"
	"github.com/spf13/cobra"
)

// newFriendsListCommand creates the friends list subcommand.
func newFriendsListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "friends-list [userID]", Short: "List friendships for a user",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			userID, err := parseID(args[0])
			if err != nil {
				return err
			}
			friends, err := service.ListFriends(context.Background(), userID)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, friends)
		}}
}

// newFriendsAddCommand creates the friends add subcommand.
func newFriendsAddCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "friends-add [userID] [friendID]", Short: "Force-add a friendship",
		Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			userID, err := parseID(args[0])
			if err != nil {
				return err
			}
			friendID, err := parseID(args[1])
			if err != nil {
				return err
			}
			if addErr := service.AddFriendship(context.Background(), userID, friendID); addErr != nil {
				return addErr
			}
			return printJSON(deps.Output, map[string]string{"status": "ok"})
		}}
}

// newFriendsRemoveCommand creates the friends remove subcommand.
func newFriendsRemoveCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "friends-remove [userID] [friendID]", Short: "Remove a friendship",
		Args: cobra.ExactArgs(2), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			userID, err := parseID(args[0])
			if err != nil {
				return err
			}
			friendID, err := parseID(args[1])
			if err != nil {
				return err
			}
			if delErr := service.RemoveFriendship(context.Background(), userID, friendID); delErr != nil {
				return delErr
			}
			return printJSON(deps.Output, map[string]string{"status": "removed"})
		}}
}

// newRequestsListCommand creates the requests list subcommand.
func newRequestsListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "requests-list [userID]", Short: "List pending friend requests",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			userID, err := parseID(args[0])
			if err != nil {
				return err
			}
			requests, err := service.ListPendingRequests(context.Background(), userID)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, requests)
		}}
}

// openService builds a minimal messenger service for CLI-only use.
func openService(opts options) (*messengerapplication.Service, func(), error) {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: opts.EnvFile, EnvPrefix: opts.EnvPrefix})
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
	repository, err := messengerstore.NewRepository(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	service, err := messengerapplication.NewService(repository, &noopSessionRegistry{}, &noopBroadcaster{}, messengerapplication.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	return service, func() { _ = sqlDB.Close() }, nil
}

// parseID parses a positive integer from a CLI argument string.
func parseID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// printJSON writes one value as JSON to w (or os.Stdout if nil).
func printJSON(w io.Writer, value any) error {
	if w == nil {
		w = os.Stdout
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(payload))
	return err
}
