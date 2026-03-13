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
