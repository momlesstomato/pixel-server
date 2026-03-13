package postgres

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
)

// TestManagerMigratesUserLoginEventsTable verifies login event schema lifecycle behavior.
func TestManagerMigratesUserLoginEventsTable(t *testing.T) {
	database := openSQLiteDatabase(t)
	manager, err := NewManager(database, Config{MigrationTable: "schema_migrations_events", SeedTable: "schema_seeds_events"}, migrations.Registry(), seeds.Registry())
	if err != nil {
		t.Fatalf("expected manager creation success, got %v", err)
	}
	if err := manager.MigrateUp(); err != nil {
		t.Fatalf("expected migration up success, got %v", err)
	}
	if !database.Migrator().HasTable(&usermodel.LoginEvent{}) {
		t.Fatalf("expected login_events table to exist")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.LoginEvent{}) {
		t.Fatalf("expected login_events table to be dropped")
	}
}
