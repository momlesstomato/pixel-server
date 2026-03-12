package postgres

import (
	"fmt"
	"strings"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	"gorm.io/gorm"
)

// Stage defines PostgreSQL startup behavior.
type Stage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializePostgreSQL creates a PostgreSQL ORM client from loaded configuration.
	InitializePostgreSQL(Config) (*gorm.DB, error)
}

// Initializer provides default PostgreSQL startup behavior.
type Initializer struct{}

// Name returns the stable initializer name.
func (initializer Initializer) Name() string {
	return "postgres"
}

// InitializePostgreSQL builds a client and optionally runs schema migrations and seeds.
func (initializer Initializer) InitializePostgreSQL(loaded Config) (*gorm.DB, error) {
	if strings.TrimSpace(loaded.DSN) == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}
	database, err := NewClient(loaded)
	if err != nil {
		return nil, err
	}
	manager, err := NewManager(database, loaded, migrations.Registry(), seeds.Registry())
	if err != nil {
		return nil, err
	}
	if loaded.MigrationAutoUp {
		if err := manager.MigrateUp(); err != nil {
			return nil, err
		}
	}
	if loaded.SeedAutoUp {
		if err := manager.SeedUp(); err != nil {
			return nil, err
		}
	}
	return database, nil
}
