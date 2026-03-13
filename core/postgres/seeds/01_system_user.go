package seeds

import (
	"errors"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step01SystemUser returns essential seed step for one default system user.
func Step01SystemUser() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260313_01_system_user",
		Migrate: func(database *gorm.DB) error {
			var row usermodel.Record
			err := database.Where("username = ?", "system").First(&row).Error
			if err == nil {
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return database.Create(&usermodel.Record{Username: "system", RealName: "System", CanChangeName: false, NoobnessLevel: 0}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("username = ?", "system").Delete(&usermodel.Record{}).Error
		},
	}
}
