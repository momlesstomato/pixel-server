package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Step10RenameCostColumns renames the legacy cost_primary, cost_secondary, and
// cost_secondary_type columns on catalog_items to names that match the Habbo
// wire-protocol field semantics: both Credits and ActivityPoints are independent
// price components — neither is "primary" or "secondary".
func Step10RenameCostColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260324_10_rename_cost_columns",
		Migrate: func(database *gorm.DB) error {
			stmts := []string{
				"ALTER TABLE catalog_items RENAME COLUMN cost_primary TO cost_credits",
				"ALTER TABLE catalog_items RENAME COLUMN cost_secondary TO cost_activity_points",
				"ALTER TABLE catalog_items RENAME COLUMN cost_secondary_type TO activity_point_type",
			}
			for _, stmt := range stmts {
				if err := database.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			stmts := []string{
				"ALTER TABLE catalog_items RENAME COLUMN cost_credits TO cost_primary",
				"ALTER TABLE catalog_items RENAME COLUMN cost_activity_points TO cost_secondary",
				"ALTER TABLE catalog_items RENAME COLUMN activity_point_type TO cost_secondary_type",
			}
			for _, stmt := range stmts {
				if err := database.Exec(stmt).Error; err != nil {
					return err
				}
			}
			return nil
		},
	}
}
