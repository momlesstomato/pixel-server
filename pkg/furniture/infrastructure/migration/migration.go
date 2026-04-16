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

// Step05AddItemPlacement returns the migration that adds placement columns to the items table.
func Step05AddItemPlacement() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260403_01_add_item_placement",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&furnituremodel.Item{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE items DROP COLUMN IF EXISTS x, DROP COLUMN IF EXISTS y, DROP COLUMN IF EXISTS z, DROP COLUMN IF EXISTS dir`).Error
		},
	}
}

// Step06AddCanLay returns the migration that adds can_lay support to item definitions.
func Step06AddCanLay() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_01_add_item_definition_can_lay",
		Migrate: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return database.AutoMigrate(&furnituremodel.Definition{})
			}
			if err := database.Exec(`ALTER TABLE item_definitions ADD COLUMN IF NOT EXISTS can_lay BOOLEAN NOT NULL DEFAULT FALSE`).Error; err != nil {
				return err
			}
			return database.Exec(`UPDATE item_definitions SET can_lay = TRUE WHERE can_lay = FALSE AND (interaction_type = 'bed' OR item_name ILIKE '%bed%' OR public_name ILIKE '%bed%')`).Error
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE item_definitions DROP COLUMN IF EXISTS can_lay`).Error
		},
	}
}

// Step07AddItemInteractionData returns the migration that adds hidden interaction and wall placement support.
func Step07AddItemInteractionData() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260409_01_add_item_interaction_data",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&furnituremodel.Item{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			return database.Exec(`ALTER TABLE items DROP COLUMN IF EXISTS interaction_data, DROP COLUMN IF EXISTS wall_position`).Error
		},
	}
}
