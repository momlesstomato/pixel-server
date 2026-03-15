package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionmodel "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/model"
	"gorm.io/gorm"
)

// Step01PermissionGroups returns migration step for permission groups schema.
func Step01PermissionGroups() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_01_permission_groups",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&permissionmodel.Group{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&permissionmodel.Group{})
		},
	}
}

// Step02GroupPermissions returns migration step for group permissions schema.
func Step02GroupPermissions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260314_02_group_permissions",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&permissionmodel.Grant{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&permissionmodel.Grant{})
		},
	}
}
