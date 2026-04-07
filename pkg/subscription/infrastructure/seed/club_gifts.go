package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	submodel "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/model"
	"gorm.io/gorm"
)

var defaultClubGifts = []submodel.ClubGift{
	{Name: "Gray Dining Chair", ItemDefinitionID: 26, DaysRequired: 31, Enabled: true, OrderNum: 1},
	{Name: "Aquamarine Double Bed", ItemDefinitionID: 41, DaysRequired: 62, Enabled: true, OrderNum: 2},
	{Name: "Telephone Box", ItemDefinitionID: 202, DaysRequired: 93, Enabled: true, OrderNum: 3},
}

var defaultPaydayConfig = submodel.PaydayConfig{ID: 1, IntervalDays: 31}

// Step03DefaultClubGifts returns seed step for essential HC monthly gifts.
func Step03DefaultClubGifts() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_S06_default_club_gifts",
		Migrate: func(database *gorm.DB) error {
			for i := range defaultClubGifts {
				if err := database.FirstOrCreate(&defaultClubGifts[i], submodel.ClubGift{Name: defaultClubGifts[i].Name}).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(defaultClubGifts))
			for _, gift := range defaultClubGifts {
				names = append(names, gift.Name)
			}
			return database.Where("name IN ?", names).Delete(&submodel.ClubGift{}).Error
		},
	}
}

// Step04DefaultPaydayConfig returns the seed step for the default HC payday configuration.
func Step04DefaultPaydayConfig() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_S07_default_payday_config",
		Migrate: func(database *gorm.DB) error {
			return database.FirstOrCreate(&defaultPaydayConfig, submodel.PaydayConfig{ID: defaultPaydayConfig.ID}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("id = ?", defaultPaydayConfig.ID).Delete(&submodel.PaydayConfig{}).Error
		},
	}
}
