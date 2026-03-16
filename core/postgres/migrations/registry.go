package migrations

import gormigrate "github.com/go-gormigrate/gormigrate/v2"
import messengermigration "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/migration"
import permissionmigration "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/migration"

// Registry returns ordered schema migration steps.
func Registry() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		permissionmigration.Step01PermissionGroups(),
		permissionmigration.Step02GroupPermissions(),
		Step01Users(),
		Step01UsersRenameLegacyRecords(),
		permissionmigration.Step03UserPermissionGroups(),
		Step02UserLoginEvents(),
		Step03UserSettings(),
		Step04UserRespects(),
		Step05UserWardrobe(),
		Step06UserIgnores(),
		messengermigration.Step01MessengerFriendships(),
		messengermigration.Step02FriendRequests(),
		messengermigration.Step03OfflineMessages(),
		messengermigration.Step04NormalizeFriendships(),
	}
}
