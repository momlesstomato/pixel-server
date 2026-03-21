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
	catalogapp "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	catalogstore "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/store"
	"github.com/spf13/cobra"
)

// newPagesListCommand creates the pages-list subcommand.
func newPagesListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "pages-list", Short: "List catalog pages",
		Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			result, err := service.ListPages(context.Background())
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newPagesGetCommand creates the pages-get subcommand.
func newPagesGetCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "pages-get [id]", Short: "Get catalog page by ID",
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
			result, err := service.FindPageByID(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newOffersListCommand creates the offers-list subcommand.
func newOffersListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "offers-list [id]", Short: "List offers by page",
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
			result, err := service.ListOffersByPageID(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// openService builds a minimal catalog service for CLI-only use.
func openService(opts options) (*catalogapp.Service, func(), error) {
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
	repository, err := catalogstore.NewRepository(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	service, err := catalogapp.NewService(repository)
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
