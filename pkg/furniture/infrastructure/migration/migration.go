package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	furnituremodel "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/model"
	"gorm.io/gorm"
)

// Step01ItemDefinitions returns the migration that creates the item_definitions table.
func Step01ItemDefinitions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_01_item_definitions",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&furnituremodel.Definition{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&furnituremodel.Definition{})
		},
	}
}

// Step03DropRevision returns the migration that removes the legacy revision
// column from item_definitions. Revision tracking is not required.
func Step03DropRevision() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_01_drop_revision",
		Migrate: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE item_definitions DROP COLUMN IF EXISTS revision`).Error
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE item_definitions ADD COLUMN IF NOT EXISTS revision INTEGER NOT NULL DEFAULT 1`).Error
		},
	}
}

// Step04RestoreSpriteID returns the migration that restores the sprite_id column
// to item_definitions after it was incorrectly dropped.
func Step04RestoreSpriteID() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260401_02_restore_sprite_id",
		Migrate: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE item_definitions ADD COLUMN IF NOT EXISTS sprite_id INTEGER NOT NULL DEFAULT 0`).Error
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE item_definitions DROP COLUMN IF EXISTS sprite_id`).Error
		},
	}
}

// Step02Items returns the migration that creates the items table.
func Step02Items() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_02_items",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&furnituremodel.Item{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&furnituremodel.Item{})
		},
	}
}
