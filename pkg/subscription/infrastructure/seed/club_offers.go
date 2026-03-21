package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

// defaultClubOffers defines essential bootstrap club membership offers.
var defaultClubOffers = []submodel.ClubOffer{
	{Name: "HC 1 Month", Days: 31, Credits: 25, Points: 0, PointsType: 0, Enabled: true},
	{Name: "HC 3 Months", Days: 93, Credits: 60, Points: 0, PointsType: 0, Enabled: true},
	{Name: "HC 12 Months", Days: 365, Credits: 200, Points: 0, PointsType: 0, Enabled: true},
}

// Step01DefaultClubOffers returns seed step for essential club membership offers.
func Step01DefaultClubOffers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_S02_club_offers",
		Migrate: func(database *gorm.DB) error {
			for i := range defaultClubOffers {
				if err := database.FirstOrCreate(&defaultClubOffers[i], submodel.ClubOffer{Name: defaultClubOffers[i].Name}).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(defaultClubOffers))
			for _, o := range defaultClubOffers {
				names = append(names, o.Name)
			}
			return database.Where("name IN ?", names).Delete(&submodel.ClubOffer{}).Error
		},
	}
}
