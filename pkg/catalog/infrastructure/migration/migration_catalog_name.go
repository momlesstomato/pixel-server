package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Step08DropCatalogName removes the catalog_name column from catalog_items.
// The display name for every offer is now always resolved from
// item_definitions.public_name via the store JOIN, eliminating
// the duplicate name storage.
func Step08DropCatalogName() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260324_08_drop_catalog_name",
		Migrate: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(
				"ALTER TABLE catalog_items DROP COLUMN IF EXISTS catalog_name",
			).Error
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(
				"ALTER TABLE catalog_items ADD COLUMN IF NOT EXISTS catalog_name VARCHAR(100) NOT NULL DEFAULT ''",
			).Error
		},
	}
}
