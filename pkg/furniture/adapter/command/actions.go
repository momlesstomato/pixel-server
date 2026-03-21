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
	furnitureapp "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furniturestore "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/store"
	"github.com/spf13/cobra"
)

// newDefinitionsListCommand creates the definitions-list subcommand.
func newDefinitionsListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "definitions-list", Short: "List item definitions",
		Args: cobra.NoArgs, RunE: func(_ *cobra.Command, _ []string) error {
			service, cleanup, err := openService(*opts)
			if err != nil {
				return err
			}
			defer cleanup()
			result, err := service.ListDefinitions(context.Background())
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newDefinitionsGetCommand creates the definitions-get subcommand.
func newDefinitionsGetCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "definitions-get [id]", Short: "Get definition by ID",
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
			result, err := service.FindDefinitionByID(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// newItemsListCommand creates the items-list subcommand.
func newItemsListCommand(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{Use: "items-list [id]", Short: "List items by owner",
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
			result, err := service.ListItemsByUserID(context.Background(), id)
			if err != nil {
				return err
			}
			return printJSON(deps.Output, result)
		}}
}

// openService builds a minimal furniture service for CLI-only use.
func openService(opts options) (*furnitureapp.Service, func(), error) {
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
	repository, err := furniturestore.NewRepository(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}
	service, err := furnitureapp.NewService(repository)
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
