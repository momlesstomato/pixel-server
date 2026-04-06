package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
)

// Step01DefaultGroups returns seed step for essential permission groups.
func Step01DefaultGroups() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_01_default_permission_groups",
		Migrate: func(database *gorm.DB) error {
			return ensureGroups(database)
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("name IN ?", []string{"default", "vip", "moderator", "admin"}).Delete(&permissionmodel.Group{}).Error
		},
	}
}

// Step03StaffAndAmbassadorGroups returns seed step for staff and ambassador permission groups.
func Step03StaffAndAmbassadorGroups() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_03_staff_ambassador_groups",
		Migrate: func(database *gorm.DB) error {
			return ensureExtendedGroups(database)
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("name IN ?", []string{"staff", "ambassador"}).Delete(&permissionmodel.Group{}).Error
		},
	}
}

// ensureGroups creates default permission groups when missing.
func ensureGroups(database *gorm.DB) error {
	defaults := []permissionmodel.Group{
		{Name: "default", DisplayName: "Default", Priority: 0, ClubLevel: 0, SecurityLevel: 0, IsAmbassador: false, IsDefault: true},
		{Name: "vip", DisplayName: "VIP", Priority: 10, ClubLevel: 2, SecurityLevel: 0, IsAmbassador: false, IsDefault: false},
		{Name: "moderator", DisplayName: "Moderator", Priority: 50, ClubLevel: 0, SecurityLevel: 1, IsAmbassador: false, IsDefault: false},
		{Name: "admin", DisplayName: "Administrator", Priority: 100, ClubLevel: 2, SecurityLevel: 3, IsAmbassador: true, IsDefault: false},
	}
	for _, group := range defaults {
		var count int64
		if err := database.Model(&permissionmodel.Group{}).Where("name = ?", group.Name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := database.Create(&group).Error; err != nil {
			return err
		}
	}
	return nil
}

// ensureExtendedGroups creates staff and ambassador groups when missing.
func ensureExtendedGroups(database *gorm.DB) error {
	extended := []permissionmodel.Group{
		{Name: "staff", DisplayName: "Staff", Priority: 75, ClubLevel: 0, SecurityLevel: 2, IsAmbassador: false, IsDefault: false},
		{Name: "ambassador", DisplayName: "Ambassador", Priority: 20, ClubLevel: 0, SecurityLevel: 0, IsAmbassador: true, IsDefault: false},
	}
	for _, group := range extended {
		var count int64
		if err := database.Model(&permissionmodel.Group{}).Where("name = ?", group.Name).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := database.Create(&group).Error; err != nil {
			return err
		}
	}
	return nil
}
