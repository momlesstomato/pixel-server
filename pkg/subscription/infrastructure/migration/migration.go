package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

// Step01Subscriptions returns the migration that creates the user_subscriptions table.
func Step01Subscriptions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_13_subscriptions",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&submodel.Subscription{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&submodel.Subscription{})
		},
	}
}
