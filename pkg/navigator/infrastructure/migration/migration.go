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

// Step05RoomPromotion returns the migration that adds promotion columns to the rooms table.
func Step05RoomPromotion() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260404_05_room_promotion",
		Migrate: func(database *gorm.DB) error {
			type rooms struct {
				PromotedUntil *string `gorm:"column:promoted_until;type:timestamptz;default:null"`
				PromotionName string  `gorm:"column:promotion_name;size:100;not null;default:''"`
			}
			return database.Table("rooms").AutoMigrate(&rooms{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			for _, col := range []string{"promoted_until", "promotion_name"} {
				if database.Migrator().HasColumn("rooms", col) {
					_ = database.Migrator().DropColumn("rooms", col)
				}
			}
			return nil
		},
	}
}

// Step06StaffPick returns the migration that adds staff_pick column to the rooms table.
func Step06StaffPick() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260404_06_staff_pick",
		Migrate: func(database *gorm.DB) error {
			type rooms struct {
				StaffPick bool `gorm:"column:staff_pick;not null;default:false"`
			}
			return database.Table("rooms").AutoMigrate(&rooms{})
		},
		Rollback: func(database *gorm.DB) error {
			if database.Dialector.Name() != "postgres" {
				return nil
			}
			if database.Migrator().HasColumn("rooms", "staff_pick") {
				return database.Migrator().DropColumn("rooms", "staff_pick")
			}
			return nil
		},
	}
}
