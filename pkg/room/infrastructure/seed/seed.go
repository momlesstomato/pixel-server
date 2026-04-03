package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	roommodel "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Step01StandardModels returns the seed that inserts predefined room model templates.
func Step01StandardModels() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "seed_20260401_01_room_models",
		Migrate: func(database *gorm.DB) error {
			models := standardModels()
			return database.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "slug"}},
				DoNothing: true,
			}).Create(&models).Error
		},
		Rollback: func(database *gorm.DB) error {
			slugs := []string{"model_a", "model_b", "model_c", "model_d", "model_e", "model_f", "model_g", "model_h", "model_i"}
			return database.Where("slug IN ?", slugs).Delete(&roommodel.RoomModel{}).Error
		},
	}
}
