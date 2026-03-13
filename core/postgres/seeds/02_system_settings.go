package seeds

import (
	"errors"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step02SystemSettings returns essential seed step for default system settings.
func Step02SystemSettings() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_02_system_settings",
		Migrate: func(database *gorm.DB) error {
			var user usermodel.Record
			if err := database.Where("username = ?", "system").First(&user).Error; err != nil {
				return err
			}
			var settings usermodel.Settings
			err := database.Where("user_id = ?", user.ID).First(&settings).Error
			if err == nil {
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return database.Create(&usermodel.Settings{UserID: user.ID}).Error
		},
		Rollback: func(database *gorm.DB) error {
			var user usermodel.Record
			if err := database.Where("username = ?", "system").First(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				return err
			}
			return database.Where("user_id = ?", user.ID).Delete(&usermodel.Settings{}).Error
		},
	}
}
