package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Step09DropCostPrimaryType removes the cost_primary_type column from
// catalog_items. The Habbo wire protocol fixes the primary cost currency
// as Credits — there is no activityPointType field on the wire for the
// primary price component. The column was unused in all encoding paths
// and is therefore removed as dead weight.
func Step09DropCostPrimaryType() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260324_09_drop_cost_primary_type",
		Migrate: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(
				"ALTER TABLE catalog_items DROP COLUMN IF EXISTS cost_primary_type",
			).Error
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(
				"ALTER TABLE catalog_items ADD COLUMN IF NOT EXISTS cost_primary_type INT NOT NULL DEFAULT 1",
			).Error
		},
	}
}
