package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	systemmodel "github.com/momlesstomato/pixel-server/core/postgres/model/system"
	"gorm.io/gorm"
)

// Step03SystemSettingsAuditOwner returns migration step for settings ownership and audit columns.
func Step03SystemSettingsAuditOwner() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260312_03_system_settings_audit_owner",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&systemmodel.Setting{})
		},
		Rollback: func(_ *gorm.DB) error {
			return nil
		},
	}
}
