package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	systemmodel "github.com/momlesstomato/pixel-server/core/postgres/model/system"
	"gorm.io/gorm"
)

const seedBootstrapVersionKey = "bootstrap_version"
const seedBootstrapVersionValue = "v1"

// Step01SystemSettingsSeed returns seed step for essential system settings.
func Step01SystemSettingsSeed() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260312_01_seed_system_settings",
		Migrate: func(database *gorm.DB) error {
			setting := systemmodel.Setting{Key: seedBootstrapVersionKey, Value: seedBootstrapVersionValue}
			return database.Where("key = ?", setting.Key).FirstOrCreate(&setting).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("key = ?", seedBootstrapVersionKey).Delete(&systemmodel.Setting{}).Error
		},
	}
}
