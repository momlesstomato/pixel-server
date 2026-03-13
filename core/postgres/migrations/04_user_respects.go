package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step04UserRespects returns migration step for user respects schema.
func Step04UserRespects() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_04_user_respects",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.Respect{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.Respect{})
		},
	}
}

// Step05UserWardrobe returns migration step for user wardrobe schema.
func Step05UserWardrobe() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_05_user_wardrobe",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.WardrobeSlot{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.WardrobeSlot{})
		},
	}
}

// Step06UserIgnores returns migration step for user ignores schema.
func Step06UserIgnores() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_06_user_ignores",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.Ignore{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.Ignore{})
		},
	}
}
