package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	permissionseed "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/seed"
)

// Registry returns ordered essential seed steps.
func Registry() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		Step01SystemUser(),
		Step02SystemSettings(),
		permissionseed.Step01DefaultGroups(),
		permissionseed.Step02DefaultPermissions(),
		Step03TestUsers(),
		Step04TestUserSettings(),
		Step05DemoUsersBackfill(),
		Step06DemoUserSettingsBackfill(),
	}
}
