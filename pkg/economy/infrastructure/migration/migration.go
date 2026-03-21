package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	economymodel "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/model"
	"gorm.io/gorm"
)

// Step01MarketplaceOffers returns the migration that creates the marketplace_offers table.
func Step01MarketplaceOffers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_10_marketplace_offers",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&economymodel.MarketplaceOffer{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&economymodel.MarketplaceOffer{})
		},
	}
}

// Step02PriceHistory returns the migration that creates the marketplace_price_history table.
func Step02PriceHistory() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_11_price_history",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&economymodel.PriceHistory{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&economymodel.PriceHistory{})
		},
	}
}

// Step03TradeLogs returns the migration that creates trade_logs and trade_log_items tables.
func Step03TradeLogs() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_12_trade_logs",
		Migrate: func(database *gorm.DB) error {
			if err := database.AutoMigrate(&economymodel.TradeLog{}); err != nil {
				return err
			}
			return database.AutoMigrate(&economymodel.TradeLogItem{})
		},
		Rollback: func(database *gorm.DB) error {
			if err := database.Migrator().DropTable(&economymodel.TradeLogItem{}); err != nil {
				return err
			}
			return database.Migrator().DropTable(&economymodel.TradeLog{})
		},
	}
}
