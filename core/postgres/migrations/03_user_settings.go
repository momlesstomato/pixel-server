package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step03UserSettings returns migration step for user settings schema.
func Step03UserSettings() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_03_user_settings",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.Settings{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.Settings{})
		},
	}
}
