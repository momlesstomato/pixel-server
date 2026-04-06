package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Step09AssignmentBackfill returns seed step that ensures every seeded user has assignment rows.
func Step09AssignmentBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_09_permission_assignment_backfill",
		Migrate: func(database *gorm.DB) error {
			if err := backfillMissingAssignments(database); err != nil {
				return err
			}
			return grantMultiGroupAssignments(database)
		},
		Rollback: func(database *gorm.DB) error {
			return nil
		},
	}
}

// backfillMissingAssignments creates assignment rows for users that have group_id but no assignments.
func backfillMissingAssignments(database *gorm.DB) error {
	type legacy struct {
		ID      uint
		GroupID uint
	}
	var rows []legacy
	if err := database.Model(&usermodel.Record{}).Select("id, group_id").Where("group_id > 0").Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		var count int64
		if err := database.Model(&permissionmodel.Assignment{}).Where("user_id = ?", row.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := database.Clauses(clause.OnConflict{DoNothing: true}).Create(&permissionmodel.Assignment{
			UserID: row.ID, GroupID: row.GroupID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// multiGroupSpec defines an extra group assignment for a user.
type multiGroupSpec struct {
	// Username stores the target user.
	Username string
	// ExtraGroups stores additional group names to assign.
	ExtraGroups []string
}

// multiGroupUsers defines users that should belong to multiple groups.
var multiGroupUsers = []multiGroupSpec{
	{Username: "test_admin", ExtraGroups: []string{"moderator", "vip"}},
	{Username: "demo_admin", ExtraGroups: []string{"moderator", "vip"}},
	{Username: "test_staff", ExtraGroups: []string{"moderator"}},
	{Username: "demo_staff", ExtraGroups: []string{"moderator"}},
}

// grantMultiGroupAssignments adds extra group assignments for selected users.
func grantMultiGroupAssignments(database *gorm.DB) error {
	allGroupNames := map[string]struct{}{}
	for _, spec := range multiGroupUsers {
		for _, name := range spec.ExtraGroups {
			allGroupNames[name] = struct{}{}
		}
	}
	names := make([]string, 0, len(allGroupNames))
	for name := range allGroupNames {
		names = append(names, name)
	}
	var groups []permissionmodel.Group
	if err := database.Where("name IN ?", names).Find(&groups).Error; err != nil {
		return err
	}
	nameToID := make(map[string]uint, len(groups))
	for _, g := range groups {
		nameToID[g.Name] = g.ID
	}
	for _, spec := range multiGroupUsers {
		var user usermodel.Record
		q := database.Where("username = ?", spec.Username).Limit(1).Find(&user)
		if q.Error != nil || q.RowsAffected == 0 {
			continue
		}
		for _, groupName := range spec.ExtraGroups {
			gID, ok := nameToID[groupName]
			if !ok {
				continue
			}
			_ = database.Clauses(clause.OnConflict{DoNothing: true}).Create(&permissionmodel.Assignment{
				UserID: user.ID, GroupID: gID,
			}).Error
		}
	}
	return nil
}
