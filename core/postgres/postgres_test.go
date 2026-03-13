package postgres

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
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
	if !database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected migrated users table to exist")
	}
	if !database.Migrator().HasTable(&usermodel.LoginEvent{}) {
		t.Fatalf("expected migrated user login events table to exist")
	}
	if !database.Migrator().HasTable(&usermodel.Settings{}) {
		t.Fatalf("expected migrated user settings table to exist")
	}
	if !database.Migrator().HasTable(&usermodel.Respect{}) {
		t.Fatalf("expected migrated user respects table to exist")
	}
	user := usermodel.Record{Username: "tester"}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("expected user insert success, got %v", err)
	}
	if user.ID == 0 || user.CreatedAt.IsZero() || user.UpdatedAt.IsZero() || user.OwnerID != nil {
		t.Fatalf("expected generated id, timestamps, and nil owner for inserted user")
	}
	if err := database.Delete(&user).Error; err != nil {
		t.Fatalf("expected user soft delete success, got %v", err)
	}
	var visibleUsers int64
	if err := database.Model(&usermodel.Record{}).Where("username = ?", user.Username).Count(&visibleUsers).Error; err != nil {
		t.Fatalf("expected visible user count query success, got %v", err)
	}
	if visibleUsers != 0 {
		t.Fatalf("expected zero visible users after soft delete, got %d", visibleUsers)
	}
	var storedUser usermodel.Record
	if err := database.Unscoped().Where("id = ?", user.ID).First(&storedUser).Error; err != nil {
		t.Fatalf("expected unscoped user lookup success, got %v", err)
	}
	if !storedUser.DeletedAt.Valid {
		t.Fatalf("expected deleted_at to be set after soft delete")
	}
	if err := manager.SeedUp(); err != nil {
		t.Fatalf("expected seed up success, got %v", err)
	}
	if err := manager.SeedDown(); err != nil {
		t.Fatalf("expected seed down success, got %v", err)
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.Respect{}) {
		t.Fatalf("expected user respects table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.Settings{}) {
		t.Fatalf("expected users settings table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.Settings{}) {
		t.Fatalf("expected user settings table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.LoginEvent{}) {
		t.Fatalf("expected user login events table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.LoginEvent{}) {
		t.Fatalf("expected user login events table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to be dropped")
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
