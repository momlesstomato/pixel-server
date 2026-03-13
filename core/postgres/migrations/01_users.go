package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step01Users returns migration step for users schema.
func Step01Users() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260312_01_users",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.Record{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.Record{})
		},
	}
}
