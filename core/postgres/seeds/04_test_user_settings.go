package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// Step04TestUserSettings returns seed step for default settings of test users.
func Step04TestUserSettings() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_04_test_user_settings",
		Migrate: func(database *gorm.DB) error {
			return migrateTestUserSettings(database)
		},
		Rollback: func(database *gorm.DB) error {
			return rollbackTestUserSettings(database)
		},
	}
}

// Step06DemoUserSettingsBackfill returns seed step to backfill demo settings on already-seeded databases.
func Step06DemoUserSettingsBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_06_demo_user_settings_backfill",
		Migrate: func(database *gorm.DB) error {
			for _, spec := range demoUserSpecs {
				if err := ensureTestUserSettings(database, spec.Username); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			for _, spec := range demoUserSpecs {
				if err := removeTestUserSettings(database, spec.Username); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// migrateTestUserSettings creates default settings rows for all test users.
func migrateTestUserSettings(database *gorm.DB) error {
	for _, spec := range testUserSpecs {
		if err := ensureTestUserSettings(database, spec.Username); err != nil {
			return err
		}
	}
	return nil
}

// ensureTestUserSettings resolves user and creates settings when missing.
func ensureTestUserSettings(database *gorm.DB, username string) error {
	var user usermodel.Record
	userQuery := database.Where("username = ?", username).Limit(1).Find(&user)
	if userQuery.Error != nil {
		return userQuery.Error
	}
	if userQuery.RowsAffected == 0 {
		return nil
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
}

// rollbackTestUserSettings removes settings for all test users.
func rollbackTestUserSettings(database *gorm.DB) error {
	for _, spec := range testUserSpecs {
		if err := removeTestUserSettings(database, spec.Username); err != nil {
			return err
		}
	}
	return nil
}

// removeTestUserSettings deletes settings row for one test user.
func removeTestUserSettings(database *gorm.DB, username string) error {
	var user usermodel.Record
	query := database.Where("username = ?", username).Limit(1).Find(&user)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected == 0 {
		return nil
	}
	return database.Where("user_id = ?", user.ID).Delete(&usermodel.Settings{}).Error
}
