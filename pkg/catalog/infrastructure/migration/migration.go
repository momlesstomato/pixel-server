package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// Step01CatalogPages returns the migration that creates the catalog_pages table.
func Step01CatalogPages() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_07_catalog_pages",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&catalogmodel.Page{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&catalogmodel.Page{})
		},
	}
}

// Step02CatalogOffers returns the migration that creates the catalog_offers table.
func Step02CatalogOffers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_08_catalog_offers",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&catalogmodel.Offer{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&catalogmodel.Offer{})
		},
	}
}

// Step04OfferCostColumns returns the migration that renames legacy cost columns on catalog_items
// to the extensible cost_primary/cost_secondary scheme and adds reward_currency_type to vouchers.
func Step04OfferCostColumns() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_08b_offer_cost_columns",
		Migrate: func(database *gorm.DB) error {
			if err := database.Exec(`
				DO $$ BEGIN
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_credits') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_credits TO cost_primary;
					END IF;
					IF NOT EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_primary') THEN
						ALTER TABLE catalog_items ADD COLUMN cost_primary INTEGER NOT NULL DEFAULT 0;
					END IF;
					IF NOT EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_primary_type') THEN
						ALTER TABLE catalog_items ADD COLUMN cost_primary_type INTEGER NOT NULL DEFAULT 1;
					END IF;
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_points') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_points TO cost_secondary;
					END IF;
					IF NOT EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_secondary') THEN
						ALTER TABLE catalog_items ADD COLUMN cost_secondary INTEGER NOT NULL DEFAULT 0;
					END IF;
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_points_type') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_points_type TO cost_secondary_type;
					END IF;
					IF NOT EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_secondary_type') THEN
						ALTER TABLE catalog_items ADD COLUMN cost_secondary_type INTEGER NOT NULL DEFAULT 0;
					END IF;
				END $$
			`).Error; err != nil {
				return err
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			return database.Exec(`
				DO $$ BEGIN
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_primary') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_primary TO cost_credits;
					END IF;
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_secondary') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_secondary TO cost_points;
					END IF;
					IF EXISTS (SELECT 1 FROM information_schema.columns
						WHERE table_name = 'catalog_items' AND column_name = 'cost_secondary_type') THEN
						ALTER TABLE catalog_items RENAME COLUMN cost_secondary_type TO cost_points_type;
					END IF;
					ALTER TABLE catalog_items DROP COLUMN IF EXISTS cost_primary_type;
				END $$
			`).Error
		},
	}
}

// Step03Vouchers returns the migration that creates vouchers and voucher_redemptions tables.
func Step03Vouchers() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_09_vouchers",
		Migrate: func(database *gorm.DB) error {
			if err := database.AutoMigrate(&catalogmodel.Voucher{}); err != nil {
				return err
			}
			return database.AutoMigrate(&catalogmodel.VoucherRedemption{})
		},
		Rollback: func(database *gorm.DB) error {
			if err := database.Migrator().DropTable(&catalogmodel.VoucherRedemption{}); err != nil {
				return err
			}
			return database.Migrator().DropTable(&catalogmodel.Voucher{})
		},
	}
}

// Step05VoucherCurrencyType returns the migration that adds reward_currency_type to vouchers.
func Step05VoucherCurrencyType() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_09b_voucher_currency_type",
		Migrate: func(database *gorm.DB) error {
			return database.Exec(
				"ALTER TABLE vouchers ADD COLUMN IF NOT EXISTS reward_currency_type INTEGER",
			).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Exec(
				"ALTER TABLE vouchers DROP COLUMN IF EXISTS reward_currency_type",
			).Error
		},
	}
}
