package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	roomstore "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/store"
	"github.com/spf13/cobra"
)

// newChatExportCommand creates the chat-export subcommand.
func newChatExportCommand(deps Dependencies, opts *options) *cobra.Command {
	var roomID int
	var dateFlag string
	cmd := &cobra.Command{Use: "chat-export", Short: "Export room chat logs as .log text",
		RunE: func(_ *cobra.Command, _ []string) error {
			if roomID <= 0 {
				return errors.New("--room-id must be a positive integer")
			}
			from, to, err := parseDateWindow(dateFlag)
			if err != nil {
				return err
			}
			store, cleanup, err := openChatLogStore(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			entries, err := store.ListByRoom(context.Background(), roomID, from, to)
			if err != nil {
				return err
			}
			w := deps.Output
			if w == nil {
				w = os.Stdout
			}
			for _, e := range entries {
				line := formatLogLine(e.CreatedAt, e.ChatType, e.Username, e.Message)
				if _, err := fmt.Fprintln(w, line); err != nil {
					return err
				}
			}
			return nil
		}}
	cmd.Flags().IntVar(&roomID, "room-id", 0, "Room identifier (required)")
	cmd.Flags().StringVar(&dateFlag, "date", "", "Date filter YYYY-MM-DD (defaults to today)")
	return cmd
}

// openChatLogStore builds a ChatLogStore from env config for CLI use.
func openChatLogStore(opts options) (*roomstore.ChatLogStore, func(), error) {
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
	store, err := roomstore.NewChatLogStore(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	return store, func() { _ = sqlDB.Close() }, nil
}

// parseDateWindow converts a date flag into a from/to time range.
func parseDateWindow(dateFlag string) (time.Time, time.Time, error) {
	if dateFlag == "" {
		now := time.Now()
		return startOfDay(now), now, nil
	}
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(dateFlag))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return startOfDay(parsed), parsed.Add(24*time.Hour - time.Nanosecond), nil
}

// startOfDay truncates a time to the start of its calendar day.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// formatLogLine formats one chat entry as a log line.
func formatLogLine(at time.Time, chatType, username, message string) string {
	ts := at.Format("15:04:05")
	ct := strings.ToUpper(chatType)
	return "[" + ts + "] [" + ct + "] " + username + ": " + message
}

// parsePositiveInt parses a positive integer from a string.
func parsePositiveInt(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("must be a positive integer")
	}
	return id, nil
}
