package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// testUserSpec defines one test user seeding specification.
type testUserSpec struct {
	// Username stores unique account name.
	Username string
	// Figure stores avatar figure string.
	Figure string
	// Gender stores avatar gender marker.
	Gender string
	// Motto stores profile motto.
	Motto string
	// GroupName stores owning permission group name.
	GroupName string
}

// demoUserSpecs defines demo users that were introduced after initial seed rollout.
var demoUserSpecs = []testUserSpec{
	{Username: "demo_default", Figure: "hr-165-45.hd-180-1.ch-210-66.lg-270-66.sh-290-80", Gender: "M", Motto: "Demo default user", GroupName: "default"},
	{Username: "demo_vip", Figure: "hr-893-45.hd-600-1.ch-3030-1408.lg-285-110.sh-295-110", Gender: "F", Motto: "Demo VIP user", GroupName: "vip"},
	{Username: "demo_moderator", Figure: "hr-828-45.hd-180-1.ch-3030-62.lg-3023-64.sh-295-64", Gender: "M", Motto: "Demo moderator user", GroupName: "moderator"},
	{Username: "demo_admin", Figure: "hr-3163-42.hd-600-1.ch-3030-82.lg-280-82.sh-290-82", Gender: "F", Motto: "Demo administrator user", GroupName: "admin"},
}

// testUserSpecs defines all test users to seed.
var testUserSpecs = []testUserSpec{
	{Username: "test_default", Figure: "hr-115-42.hd-180-1.ch-3030-82.lg-275-82.sh-295-62", Gender: "M", Motto: "Default test user", GroupName: "default"},
	{Username: "test_vip", Figure: "hr-3163-42.hd-180-1.ch-3030-62.lg-3058-62.sh-295-62", Gender: "F", Motto: "VIP test user", GroupName: "vip"},
	{Username: "test_moderator", Figure: "hr-3322-1304.hd-180-1.ch-3030-92.lg-3058-1408.sh-295-1408", Gender: "M", Motto: "Moderator test user", GroupName: "moderator"},
	{Username: "test_admin", Figure: "hr-3163-1304.hd-180-1.ch-3030-1408.lg-275-1408.sh-295-1408", Gender: "M", Motto: "Admin test user", GroupName: "admin"},
	demoUserSpecs[0],
	demoUserSpecs[1],
	demoUserSpecs[2],
	demoUserSpecs[3],
}

// Step03TestUsers returns seed step for test user accounts.
func Step03TestUsers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_03_test_users",
		Migrate: func(database *gorm.DB) error {
			return migrateTestUsers(database)
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(testUserSpecs))
			for _, spec := range testUserSpecs {
				names = append(names, spec.Username)
			}
			return database.Unscoped().Where("username IN ?", names).Delete(&usermodel.Record{}).Error
		},
	}
}

// Step05DemoUsersBackfill returns seed step to backfill demo users on already-seeded databases.
func Step05DemoUsersBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_05_demo_users_backfill",
		Migrate: func(database *gorm.DB) error {
			groupIDs, err := resolveGroupIDs(database)
			if err != nil {
				return err
			}
			for _, spec := range demoUserSpecs {
				if err := ensureTestUser(database, spec, groupIDs); err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			names := make([]string, 0, len(demoUserSpecs))
			for _, spec := range demoUserSpecs {
				names = append(names, spec.Username)
			}
			return database.Unscoped().Where("username IN ?", names).Delete(&usermodel.Record{}).Error
		},
	}
}

// migrateTestUsers resolves group identifiers and creates missing test users.
func migrateTestUsers(database *gorm.DB) error {
	groupIDs, err := resolveGroupIDs(database)
	if err != nil {
		return err
	}
	for _, spec := range testUserSpecs {
		if err := ensureTestUser(database, spec, groupIDs); err != nil {
			return err
		}
	}
	return nil
}

// resolveGroupIDs returns a map of group name to identifier.
func resolveGroupIDs(database *gorm.DB) (map[string]uint, error) {
	names := []string{"default", "vip", "moderator", "admin"}
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

// ensureTestUser creates one test user when missing.
func ensureTestUser(database *gorm.DB, spec testUserSpec, groupIDs map[string]uint) error {
	var row usermodel.Record
	query := database.Where("username = ?", spec.Username).Limit(1).Find(&row)
	if query.Error != nil {
		return query.Error
	}
	if query.RowsAffected > 0 {
		return nil
	}
	groupID := groupIDs[spec.GroupName]
	return database.Create(&usermodel.Record{
		Username: spec.Username, Figure: spec.Figure, Gender: spec.Gender,
		Motto: spec.Motto, CanChangeName: true, NoobnessLevel: 0, GroupID: groupID,
	}).Error
}
