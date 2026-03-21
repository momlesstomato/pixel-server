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
