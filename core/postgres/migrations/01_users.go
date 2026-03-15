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

// Step01UsersRenameLegacyRecords returns migration step for records->users table migration.
func Step01UsersRenameLegacyRecords() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_04_users_table_rename",
		Migrate: func(database *gorm.DB) error {
			if database.Migrator().HasTable(&usermodel.Record{}) {
				return nil
			}
			if !database.Migrator().HasTable("records") {
				return nil
			}
			return database.Exec("ALTER TABLE records RENAME TO users").Error
		},
		Rollback: func(database *gorm.DB) error {
			return nil
		},
	}
}
