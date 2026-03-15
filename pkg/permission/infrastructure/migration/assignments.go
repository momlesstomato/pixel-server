package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	usermodel "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Step03UserPermissionGroups returns migration step for user group assignment schema.
func Step03UserPermissionGroups() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_03_user_permission_groups",
		Migrate: func(database *gorm.DB) error {
			if err := database.AutoMigrate(&permissionmodel.Assignment{}); err != nil {
				return err
			}
			return backfillLegacyUserGroups(database)
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&permissionmodel.Assignment{})
		},
	}
}

// backfillLegacyUserGroups migrates users.group_id values into assignment rows.
func backfillLegacyUserGroups(database *gorm.DB) error {
	type legacy struct {
		ID      uint
		GroupID uint
	}
	var rows []legacy
	if !database.Migrator().HasTable(&usermodel.Record{}) || !database.Migrator().HasColumn(&usermodel.Record{}, "group_id") {
		return nil
	}
	if err := database.Model(&usermodel.Record{}).Select("id, group_id").Where("group_id > 0").Scan(&rows).Error; err != nil {
		return err
	}
	assignments := make([]permissionmodel.Assignment, 0, len(rows))
	for _, row := range rows {
		assignments = append(assignments, permissionmodel.Assignment{UserID: row.ID, GroupID: row.GroupID})
	}
	if len(assignments) == 0 {
		return nil
	}
	return database.Clauses(clause.OnConflict{DoNothing: true}).Create(&assignments).Error
}
