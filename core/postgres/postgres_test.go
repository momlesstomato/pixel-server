package postgres

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	systemmodel "github.com/momlesstomato/pixel-server/core/postgres/model/system"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewClientRejectsMissingDSN verifies client precondition validation.
func TestNewClientRejectsMissingDSN(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatalf("expected postgres client creation failure for missing dsn")
	}
}

// TestManagerMigrateSeedUpDownWithDefaults verifies default migration and seed lifecycle behavior.
func TestManagerMigrateSeedUpDownWithDefaults(t *testing.T) {
	database := openSQLiteDatabase(t)
	manager, err := NewManager(database, Config{MigrationTable: "schema_migrations_test", SeedTable: "schema_seeds_test"}, migrations.Registry(), seeds.Registry())
	if err != nil {
		t.Fatalf("expected manager creation success, got %v", err)
	}
	if err := manager.MigrateUp(); err != nil {
		t.Fatalf("expected migration up success, got %v", err)
	}
	if !database.Migrator().HasTable(&systemmodel.Setting{}) {
		t.Fatalf("expected migrated settings table to exist")
	}
	if err := manager.SeedUp(); err != nil {
		t.Fatalf("expected seed up success, got %v", err)
	}
	var seededCount int64
	if err := database.Model(&systemmodel.Setting{}).Where("key = ?", "bootstrap_version").Count(&seededCount).Error; err != nil {
		t.Fatalf("expected seeded setting count query success, got %v", err)
	}
	if seededCount != 1 {
		t.Fatalf("expected one seeded setting, got %d", seededCount)
	}
	if err := manager.SeedDown(); err != nil {
		t.Fatalf("expected seed down success, got %v", err)
	}
	if err := database.Model(&systemmodel.Setting{}).Where("key = ?", "bootstrap_version").Count(&seededCount).Error; err != nil {
		t.Fatalf("expected unseed count query success, got %v", err)
	}
	if seededCount != 0 {
		t.Fatalf("expected zero seeded settings after rollback, got %d", seededCount)
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&systemmodel.Setting{}) {
		t.Fatalf("expected settings table to be dropped")
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
	if _, err := NewManager(database, Config{}, migrations.Registry(), nil); err == nil {
		t.Fatalf("expected empty seeder validation failure")
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
