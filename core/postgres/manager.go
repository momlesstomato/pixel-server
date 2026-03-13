package postgres

import (
	"fmt"
	"strings"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Manager defines migration and seed lifecycle behavior.
type Manager struct {
	// migrator stores schema migration state machine.
	migrator *gormigrate.Gormigrate
	// seeder stores essential seed state machine.
	seeder *gormigrate.Gormigrate
}

// NewManager creates a migration manager for schema and seeds.
func NewManager(database *gorm.DB, loaded Config, migrations []*gormigrate.Migration, seeders []*gormigrate.Migration) (*Manager, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	if len(migrations) == 0 {
		return nil, fmt.Errorf("at least one migration is required")
	}
	migrationTable := strings.TrimSpace(loaded.MigrationTable)
	if migrationTable == "" {
		migrationTable = "schema_migrations"
	}
	migrator := gormigrate.New(database, &gormigrate.Options{TableName: migrationTable}, migrations)
	var seeder *gormigrate.Gormigrate
	if len(seeders) > 0 {
		seedTable := strings.TrimSpace(loaded.SeedTable)
		if seedTable == "" {
			seedTable = "schema_seeds"
		}
		seeder = gormigrate.New(database, &gormigrate.Options{TableName: seedTable}, seeders)
	}
	return &Manager{migrator: migrator, seeder: seeder}, nil
}

// MigrateUp applies all pending schema migrations.
func (manager *Manager) MigrateUp() error { return manager.migrator.Migrate() }

// MigrateDown rolls back the last applied schema migration.
func (manager *Manager) MigrateDown() error { return manager.migrator.RollbackLast() }

// SeedUp applies all pending essential seed units.
func (manager *Manager) SeedUp() error {
	if manager.seeder == nil {
		return nil
	}
	return manager.seeder.Migrate()
}

// SeedDown rolls back the last applied essential seed unit.
func (manager *Manager) SeedDown() error {
	if manager.seeder == nil {
		return nil
	}
	return manager.seeder.RollbackLast()
}
