package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	catalogseed "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/seed"
	inventoryseed "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/seed"
	navigatorseed "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/seed"
	permissionseed "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/seed"
	subscriptionseed "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/seed"
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
		inventoryseed.Step01CurrencyTypes(),
		catalogseed.Step01DefaultPages(),
		subscriptionseed.Step01DefaultClubOffers(),
		navigatorseed.Step01DefaultCategories(),
		navigatorseed.Step02DemoRooms(),
	}
}
