package postgres

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	for i := range 4 {
		if err := manager.MigrateDown(); err != nil {
			t.Fatalf("rollback messenger step %d: %v", 4-i, err)
		}
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

// TestNewManagerRejectsMissingInputs verifies manager constructor validation behavior.
func TestNewManagerRejectsMissingInputs(t *testing.T) {
	database := openSQLiteDatabase(t)
	if _, err := NewManager(nil, Config{}, migrations.Registry(), seeds.Registry()); err == nil {
		t.Fatalf("expected nil database validation failure")
	}
	if _, err := NewManager(database, Config{}, nil, seeds.Registry()); err == nil {
		t.Fatalf("expected empty migration validation failure")
	}
	if _, err := NewManager(database, Config{}, migrations.Registry(), nil); err != nil {
		t.Fatalf("expected nil seeder list acceptance, got %v", err)
	}
}

// TestInitializerRejectsMissingDSN verifies initializer precondition checks.
func TestInitializerRejectsMissingDSN(t *testing.T) {
	if _, err := (Initializer{}).InitializePostgreSQL(Config{}); err == nil {
		t.Fatalf("expected initializer failure for missing dsn")
	}
}

// openSQLiteDatabase creates a gorm sqlite database for migration lifecycle tests.
func openSQLiteDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite database creation success, got %v", err)
	}
	return database
}
