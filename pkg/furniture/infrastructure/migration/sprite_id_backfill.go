package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Step08BackfillZeroSpriteID returns the migration that repairs zero-valued sprite_id rows by restoring the definition id mapping. Rollback is a no-op because zero sprite identifiers are invalid and cannot be reconstructed safely without external source data.
func Step08BackfillZeroSpriteID() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260415_01_backfill_item_definition_sprite_id",
		Migrate: func(database *gorm.DB) error {
			return database.Exec(`UPDATE item_definitions SET sprite_id = id WHERE sprite_id = 0`).Error
		},
		Rollback: func(_ *gorm.DB) error {
			return nil
		},
	}
}