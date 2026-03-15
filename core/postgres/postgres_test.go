package postgres

import (
	"testing"

	"github.com/momlesstomato/pixel-server/core/postgres/migrations"
	"github.com/momlesstomato/pixel-server/core/postgres/seeds"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
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
	if !database.Migrator().HasTable(&usermodel.WardrobeSlot{}) {
		t.Fatalf("expected migrated user wardrobe table to exist")
	}
	if !database.Migrator().HasTable(&usermodel.Ignore{}) {
		t.Fatalf("expected migrated user ignores table to exist")
	}
	if !database.Migrator().HasTable(&permissionmodel.Group{}) {
		t.Fatalf("expected migrated permission groups table to exist")
	}
	if !database.Migrator().HasTable(&permissionmodel.Grant{}) {
		t.Fatalf("expected migrated group permissions table to exist")
	}
	if !database.Migrator().HasTable(&permissionmodel.Assignment{}) {
		t.Fatalf("expected migrated user permission groups table to exist")
	}
	assertUserSoftDeleteLifecycle(t, database)
	if err := manager.SeedUp(); err != nil {
		t.Fatalf("expected seed up success, got %v", err)
	}
	if err := manager.SeedDown(); err != nil {
		t.Fatalf("expected seed down success, got %v", err)
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.Ignore{}) {
		t.Fatalf("expected user ignores table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.WardrobeSlot{}) {
		t.Fatalf("expected user wardrobe table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.WardrobeSlot{}) {
		t.Fatalf("expected user wardrobe table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.Respect{}) {
		t.Fatalf("expected user respects table to remain")
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
	if !database.Migrator().HasTable(&permissionmodel.Assignment{}) {
		t.Fatalf("expected user permission groups table to remain")
	}
	if !database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&permissionmodel.Assignment{}) {
		t.Fatalf("expected user permission groups table to be dropped")
	}
	if !database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if !database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to remain after transitional rename rollback step")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&usermodel.Record{}) {
		t.Fatalf("expected users table to be dropped")
	}
	if !database.Migrator().HasTable(&permissionmodel.Grant{}) {
		t.Fatalf("expected group permissions table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&permissionmodel.Grant{}) {
		t.Fatalf("expected group permissions table to be dropped")
	}
	if !database.Migrator().HasTable(&permissionmodel.Group{}) {
		t.Fatalf("expected permission groups table to remain")
	}
	if err := manager.MigrateDown(); err != nil {
		t.Fatalf("expected migration down success, got %v", err)
	}
	if database.Migrator().HasTable(&permissionmodel.Group{}) {
		t.Fatalf("expected permission groups table to be dropped")
	}
}
