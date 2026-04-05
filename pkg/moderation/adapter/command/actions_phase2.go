package command

import (
	"context"
	"fmt"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	"github.com/momlesstomato/pixel-server/pkg/moderation/infrastructure/store"
	"github.com/spf13/cobra"
)

// openDB loads config and opens a postgres connection.
func openDB(opts *options) (*store.WordFilterStore, *store.PresetStore, func(), error) {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: opts.envFile, EnvPrefix: opts.envPrefix})
	if err != nil {
		return nil, nil, nil, err
	}
	database, err := postgrescore.NewClient(loaded.PostgreSQL)
	if err != nil {
		return nil, nil, nil, err
	}
	sqlDB, err := database.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	wf, err := store.NewWordFilterStore(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, nil, err
	}
	ps, err := store.NewPresetStore(database)
	if err != nil {
		_ = sqlDB.Close()
		return nil, nil, nil, err
	}
	return wf, ps, func() { _ = sqlDB.Close() }, nil
}

// newWordFilterCommand creates the wordfilter subcommand tree.
func newWordFilterCommand(deps Dependencies, opts *options) *cobra.Command {
	cmd := &cobra.Command{Use: "wordfilter", Short: "Manage word filters"}
	cmd.AddCommand(newWordFilterListCmd(deps, opts))
	cmd.AddCommand(newWordFilterDeleteCmd(deps, opts))
	return cmd
}

// newWordFilterListCmd creates the wordfilter list subcommand.
func newWordFilterListCmd(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List active word filters",
		RunE: func(_ *cobra.Command, _ []string) error {
			wf, _, cleanup, err := openDB(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			filters, err := wf.ListActive(context.Background(), "", 0)
			if err != nil {
				return err
			}
			for _, f := range filters {
				fmt.Fprintf(deps.Output, "[%d] pattern=%q replacement=%q regex=%t scope=%s\n",
					f.ID, f.Pattern, f.Replacement, f.IsRegex, f.Scope)
			}
			if len(filters) == 0 {
				fmt.Fprintln(deps.Output, "No active word filters found.")
			}
			return nil
		},
	}
}

// newWordFilterDeleteCmd creates the wordfilter delete subcommand.
func newWordFilterDeleteCmd(deps Dependencies, opts *options) *cobra.Command {
	var filterID int64
	cmd := &cobra.Command{
		Use: "delete", Short: "Delete a word filter",
		RunE: func(_ *cobra.Command, _ []string) error {
			if filterID <= 0 {
				return fmt.Errorf("--filter-id is required")
			}
			wf, _, cleanup, err := openDB(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			if err := wf.Delete(context.Background(), filterID); err != nil {
				return err
			}
			fmt.Fprintf(deps.Output, "Deleted word filter #%d\n", filterID)
			return nil
		},
	}
	cmd.Flags().Int64Var(&filterID, "filter-id", 0, "Filter ID to delete")
	return cmd
}

// newPresetCommand creates the preset subcommand tree.
func newPresetCommand(deps Dependencies, opts *options) *cobra.Command {
	cmd := &cobra.Command{Use: "preset", Short: "Manage moderation presets"}
	cmd.AddCommand(newPresetListCmd(deps, opts))
	cmd.AddCommand(newPresetDeleteCmd(deps, opts))
	return cmd
}

// newPresetListCmd creates the preset list subcommand.
func newPresetListCmd(deps Dependencies, opts *options) *cobra.Command {
	return &cobra.Command{
		Use: "list", Short: "List active presets",
		RunE: func(_ *cobra.Command, _ []string) error {
			_, ps, cleanup, err := openDB(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			presets, err := ps.ListActive(context.Background())
			if err != nil {
				return err
			}
			for _, p := range presets {
				fmt.Fprintf(deps.Output, "[%d] %s/%s action=%s duration=%dm\n",
					p.ID, p.Category, p.Name, p.ActionType, p.DefaultDuration)
			}
			if len(presets) == 0 {
				fmt.Fprintln(deps.Output, "No active presets found.")
			}
			return nil
		},
	}
}

// newPresetDeleteCmd creates the preset delete subcommand.
func newPresetDeleteCmd(deps Dependencies, opts *options) *cobra.Command {
	var presetID int64
	cmd := &cobra.Command{
		Use: "delete", Short: "Delete a moderation preset",
		RunE: func(_ *cobra.Command, _ []string) error {
			if presetID <= 0 {
				return fmt.Errorf("--preset-id is required")
			}
			_, ps, cleanup, err := openDB(opts)
			if err != nil {
				return err
			}
			defer cleanup()
			if err := ps.Delete(context.Background(), presetID); err != nil {
				return err
			}
			fmt.Fprintf(deps.Output, "Deleted preset #%d\n", presetID)
			return nil
		},
	}
	cmd.Flags().Int64Var(&presetID, "preset-id", 0, "Preset ID to delete")
	return cmd
}
