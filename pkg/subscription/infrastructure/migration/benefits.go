package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

// Step03Benefits returns the migration that creates payday config, benefits state, and club gift tables.
func Step03Benefits() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_15_subscription_benefits",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&submodel.PaydayConfig{}, &submodel.BenefitsState{}, &submodel.ClubGift{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&submodel.ClubGift{}, &submodel.BenefitsState{}, &submodel.PaydayConfig{})
		},
	}
}
