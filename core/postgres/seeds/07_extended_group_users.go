package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// extendedUserSpecs defines test users for the staff and ambassador groups.
var extendedUserSpecs = []testUserSpec{
	{Username: "test_staff", Figure: "hr-828-45.hd-180-1.ch-3030-62.lg-3023-64.sh-295-64", Gender: "M", Motto: "Staff test user", GroupName: "staff"},
	{Username: "test_ambassador", Figure: "hr-3163-42.hd-600-1.ch-3030-82.lg-280-82.sh-290-82", Gender: "F", Motto: "Ambassador test user", GroupName: "ambassador"},
	{Username: "demo_staff", Figure: "hr-828-45.hd-180-1.ch-3030-62.lg-3023-64.sh-295-64", Gender: "M", Motto: "Demo staff user", GroupName: "staff"},
	{Username: "demo_ambassador", Figure: "hr-3163-42.hd-600-1.ch-3030-82.lg-280-82.sh-290-82", Gender: "F", Motto: "Demo ambassador user", GroupName: "ambassador"},
}

// Step07ExtendedGroupUsers returns seed step for staff and ambassador test users.
func Step07ExtendedGroupUsers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_07_extended_group_users",
		Migrate: func(database *gorm.DB) error {
			return migrateExtendedGroupUsers(database)
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(extendedUserSpecs))
			for _, spec := range extendedUserSpecs {
				names = append(names, spec.Username)
			}
			return database.Unscoped().Where("username IN ?", names).Delete(&usermodel.Record{}).Error
		},
	}
}

// Step08ExtendedGroupUserSettings returns seed step for settings of staff and ambassador users.
func Step08ExtendedGroupUserSettings() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_08_extended_group_user_settings",
		Migrate: func(database *gorm.DB) error {
			for _, spec := range extendedUserSpecs {
				if err := ensureTestUserSettings(database, spec.Username); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			for _, spec := range extendedUserSpecs {
				if err := removeTestUserSettings(database, spec.Username); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// migrateExtendedGroupUsers resolves group identifiers and creates missing extended users.
func migrateExtendedGroupUsers(database *gorm.DB) error {
	groupIDs, err := resolveExtendedGroupIDs(database)
	if err != nil {
		return err
	}
	for _, spec := range extendedUserSpecs {
		if err := ensureTestUser(database, spec, groupIDs); err != nil {
			return err
		}
	}
	return nil
}

// resolveExtendedGroupIDs returns a map of group name to identifier for staff and ambassador groups.
func resolveExtendedGroupIDs(database *gorm.DB) (map[string]uint, error) {
	names := []string{"staff", "ambassador"}
	var groups []permissionmodel.Group
	if err := database.Where("name IN ?", names).Find(&groups).Error; err != nil {
		return nil, err
	}
	result := make(map[string]uint, len(groups))
	for _, group := range groups {
		result[group.Name] = group.ID
	}
	return result, nil
}
