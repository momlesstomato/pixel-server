package seeds

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	catalogseed "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/seed"
	inventoryseed "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/seed"
	navigatorseed "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/seed"
	permissionseed "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/seed"
	roomseed "github.com/momlesstomato/pixel-server/pkg/room/infrastructure/seed"
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
		permissionseed.Step03StaffAndAmbassadorGroups(),
		permissionseed.Step04StaffAndAmbassadorPermissions(),
		Step07ExtendedGroupUsers(),
		Step08ExtendedGroupUserSettings(),
		permissionseed.Step05SecurityLevelBackfill(),
		inventoryseed.Step01CurrencyTypes(),
		catalogseed.Step01DefaultPages(),
		catalogseed.Step02HCShopPage(),
		catalogseed.Step03HCShopLocalizationBackfill(),
		catalogseed.Step04ClubGiftsPage(),
		catalogseed.Step05HCShopVipBuyBackfill(),
		subscriptionseed.Step01DefaultClubOffers(),
		subscriptionseed.Step02SubscriptionUsers(),
		subscriptionseed.Step03DefaultClubGifts(),
		subscriptionseed.Step04DefaultPaydayConfig(),
		navigatorseed.Step01DefaultCategories(),
		navigatorseed.Step02DemoRooms(),
		navigatorseed.Step03DemoAdminRoomOwnerBackfill(),
		roomseed.Step01StandardModels(),
		Step09AssignmentBackfill(),
	}
}
