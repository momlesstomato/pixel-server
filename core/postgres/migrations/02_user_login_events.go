package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step02UserLoginEvents returns migration step for user login event schema.
func Step02UserLoginEvents() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260312_02_user_login_events",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.LoginEvent{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.LoginEvent{})
		},
	}
}
