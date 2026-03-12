package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/core/postgres/model/user"
	"gorm.io/gorm"
)

// Step02Users returns migration step for users schema.
func Step02Users() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260312_02_users",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&usermodel.Record{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&usermodel.Record{})
		},
	}
}
