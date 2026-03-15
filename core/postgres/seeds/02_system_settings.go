package seeds

import (
	"fmt"

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
			userQuery := database.Where("username = ?", "system").Limit(1).Find(&user)
			if userQuery.Error != nil {
				return userQuery.Error
			}
			if userQuery.RowsAffected == 0 {
				return fmt.Errorf("system user is required before system settings")
			}
			var settings usermodel.Settings
			settingsQuery := database.Where("user_id = ?", user.ID).Limit(1).Find(&settings)
			if settingsQuery.Error != nil {
				return settingsQuery.Error
			}
			if settingsQuery.RowsAffected > 0 {
				return nil
			}
			return database.Create(&usermodel.Settings{UserID: user.ID}).Error
		},
		Rollback: func(database *gorm.DB) error {
			var user usermodel.Record
			query := database.Where("username = ?", "system").Limit(1).Find(&user)
			if query.Error != nil {
				return query.Error
			}
			if query.RowsAffected == 0 {
				return nil
			}
			return database.Where("user_id = ?", user.ID).Delete(&usermodel.Settings{}).Error
		},
	}
}
