package seeds

import (
	"errors"

	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
)

// testUserSpec defines one test user record to seed.
type testUserSpec struct {
	// Username stores unique user name value.
	Username string
	// RealName stores display real name.
	RealName string
	// GroupName stores the permission group name to resolve.
	GroupName string
}

// Step03TestUsers returns essential seed step for bootstrap test users.
func Step03TestUsers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260315_03_test_users",
		Migrate: func(database *gorm.DB) error {
			return ensureTestUsers(database)
		},
		Rollback: func(database *gorm.DB) error {
			names := []string{"alice", "bob", "charlie", "dave"}
			return database.Where("username IN ?", names).Delete(&usermodel.Record{}).Error
		},
	}
}

// ensureTestUsers creates bootstrap test users with appropriate permission groups.
func ensureTestUsers(database *gorm.DB) error {
	specs := []testUserSpec{
		{Username: "alice", RealName: "Alice", GroupName: "default"},
		{Username: "bob", RealName: "Bob", GroupName: "vip"},
		{Username: "charlie", RealName: "Charlie", GroupName: "moderator"},
		{Username: "dave", RealName: "Dave", GroupName: "admin"},
	}
	groupIDs, err := resolveGroupIDs(database, specs)
	if err != nil {
		return err
	}
	for _, spec := range specs {
		if err := ensureTestUser(database, spec, groupIDs[spec.GroupName]); err != nil {
			return err
		}
	}
	return nil
}

// resolveGroupIDs loads group identifiers by name for each test user spec.
func resolveGroupIDs(database *gorm.DB, specs []testUserSpec) (map[string]uint, error) {
	names := make([]string, 0, len(specs))
	for _, spec := range specs {
		names = append(names, spec.GroupName)
	}
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

// ensureTestUser creates one test user record when not already present.
func ensureTestUser(database *gorm.DB, spec testUserSpec, groupID uint) error {
	var row usermodel.Record
	err := database.Where("username = ?", spec.Username).First(&row).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return database.Create(&usermodel.Record{
		Username: spec.Username, RealName: spec.RealName,
		CanChangeName: true, NoobnessLevel: 0, GroupID: groupID,
	}).Error
}
