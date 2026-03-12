package cli

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/config"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	"github.com/spf13/cobra"
)

// DBOptions defines configuration inputs for database command execution.
type DBOptions struct {
	// EnvFile defines the config file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
}

// NewDBCommand creates database migration and seed command tree.
func NewDBCommand() *cobra.Command {
	var options DBOptions
	command := &cobra.Command{Use: "db", Short: "Database migration and seeding operations"}
	command.PersistentFlags().StringVar(&options.EnvFile, "env-file", ".env", "Environment file path")
	command.PersistentFlags().StringVar(&options.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.AddCommand(newDBActionCommand("migrate-up", "Apply all pending schema migrations", options, "migrate-up"))
	command.AddCommand(newDBActionCommand("migrate-down", "Rollback the last schema migration", options, "migrate-down"))
	command.AddCommand(newDBActionCommand("seed-up", "Apply all pending essential seeds", options, "seed-up"))
	command.AddCommand(newDBActionCommand("seed-down", "Rollback the last essential seed", options, "seed-down"))
	return command
}

// newDBActionCommand creates one concrete database action command.
func newDBActionCommand(use string, short string, options DBOptions, action string) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, RunE: func(_ *cobra.Command, _ []string) error { return executeDBAction(options, action) }}
}

// executeDBAction resolves configuration and runs one migration/seed action.
func executeDBAction(options DBOptions, action string) error {
	loaded, err := config.Load(config.LoaderOptions{EnvFile: options.EnvFile, EnvPrefix: options.EnvPrefix})
	if err != nil {
		return err
	}
	database, err := postgrescore.NewClient(loaded.PostgreSQL)
	if err != nil {
		return err
	}
	sqlDatabase, dbErr := database.DB()
	if dbErr == nil {
		defer sqlDatabase.Close()
	}
	manager, err := postgrescore.NewManager(database, loaded.PostgreSQL, migrations.Registry(), seeds.Registry())
	if err != nil {
		return err
	}
	switch action {
	case "migrate-up":
		return manager.MigrateUp()
	case "migrate-down":
		return manager.MigrateDown()
	case "seed-up":
		return manager.SeedUp()
	case "seed-down":
		return manager.SeedDown()
	default:
		return fmt.Errorf("unsupported db action %q", action)
	}
}
