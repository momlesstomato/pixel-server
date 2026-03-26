package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

// Step02ClubOffers returns the migration that creates the catalog_club_offers table.
func Step02ClubOffers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260325_14_club_offers",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&submodel.ClubOffer{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&submodel.ClubOffer{})
		},
	}
}
