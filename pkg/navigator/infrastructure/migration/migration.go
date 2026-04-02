package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	navmodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	"gorm.io/gorm"
)

// Step01NavigatorCategories returns the migration that creates the navigator_categories table.
func Step01NavigatorCategories() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260325_13_navigator_categories",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&navmodel.Category{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&navmodel.Category{})
		},
	}
}

// Step02Rooms returns the migration that creates the rooms table.
func Step02Rooms() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260325_14_rooms",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&navmodel.Room{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&navmodel.Room{})
		},
	}
}

// Step03SavedSearches returns the migration that creates the navigator_saved_searches table.
func Step03SavedSearches() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260325_15_navigator_saved_searches",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&navmodel.SavedSearch{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&navmodel.SavedSearch{})
		},
	}
}

// Step04Favourites returns the migration that creates the navigator_favourites table.
func Step04Favourites() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260325_16_navigator_favourites",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&navmodel.Favourite{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&navmodel.Favourite{})
		},
	}
}
