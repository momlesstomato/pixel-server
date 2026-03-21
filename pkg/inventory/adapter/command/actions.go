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
	inventoryapp "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	inventorystore "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/store"
	"github.com/spf13/cobra"
)

// newCreditsGetCommand creates the credits-get subcommand.
func newCreditsGetCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "credits-get [id]", Short: "Get user credits",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			result, err := service.GetCredits(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newCurrenciesListCommand creates the currencies-list subcommand.
func newCurrenciesListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "currencies-list [id]", Short: "List user currencies",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			result, err := service.ListCurrencies(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newBadgesListCommand creates the badges-list subcommand.
func newBadgesListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "badges-list [id]", Short: "List user badges",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			result, err := service.ListBadges(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newEffectsListCommand creates the effects-list subcommand.
func newEffectsListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "effects-list [id]", Short: "List user effects",
		Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			id, err := parseID(args[0])
			if err != nil {
				return err
			}
			result, err := service.ListEffects(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// openService builds a minimal inventory service for CLI-only use.
func openService(opts options) (*inventoryapp.Service, func(), error) {
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
	repository, err := inventorystore.NewRepository(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	service, err := inventoryapp.NewService(repository)
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
