package seeds

import (
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
			query := database.Where("username = ?", "system").Limit(1).Find(&row)
			if query.Error != nil {
				return query.Error
			}
			if query.RowsAffected > 0 {
				return nil
			}
			return database.Create(&usermodel.Record{Username: "system", RealName: "System", CanChangeName: false, NoobnessLevel: 0}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("username = ?", "system").Delete(&usermodel.Record{}).Error
		},
	}
}
