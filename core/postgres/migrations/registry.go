package migrations

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	catalogmigration "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/migration"
	economymigration "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/migration"
	furnituremigration "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/migration"
	inventorymigration "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/migration"
	messengermigration "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/migration"
	permissionmigration "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/migration"
	subscriptionmigration "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/migration"
)

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
		furnituremigration.Step01ItemDefinitions(),
		furnituremigration.Step02Items(),
		inventorymigration.Step01UserCurrencies(),
		inventorymigration.Step02DropUserCredits(),
		inventorymigration.Step03UserBadges(),
		inventorymigration.Step04UserEffects(),
		catalogmigration.Step01CatalogPages(),
		catalogmigration.Step02CatalogOffers(),
		catalogmigration.Step04OfferCostColumns(),
		catalogmigration.Step03Vouchers(),
		catalogmigration.Step05VoucherCurrencyType(),
		economymigration.Step01MarketplaceOffers(),
		economymigration.Step02PriceHistory(),
		economymigration.Step03TradeLogs(),
		subscriptionmigration.Step01Subscriptions(),
	}
}
