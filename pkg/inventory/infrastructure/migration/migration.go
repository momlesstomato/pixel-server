package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// Step01UserCurrencies returns the migration that creates currency_types, user_currencies,
// and currency_transactions tables.
func Step01UserCurrencies() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_03_user_currencies",
		Migrate: func(database *gorm.DB) error {
			if err := database.AutoMigrate(&inventorymodel.CurrencyType{}); err != nil {
				return err
			}
			if err := database.AutoMigrate(&inventorymodel.Currency{}); err != nil {
				return err
			}
			return database.AutoMigrate(&inventorymodel.CurrencyTransaction{})
		},
		Rollback: func(database *gorm.DB) error {
			if err := database.Migrator().DropTable(&inventorymodel.CurrencyTransaction{}); err != nil {
				return err
			}
			if err := database.Migrator().DropTable(&inventorymodel.Currency{}); err != nil {
				return err
			}
			return database.Migrator().DropTable(&inventorymodel.CurrencyType{})
		},
	}
}

// Step02DropUserCredits returns the migration that removes the legacy users.credits column
// if it exists, as credits are now stored in user_currencies.
func Step02DropUserCredits() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_04_drop_user_credits",
		Migrate: func(database *gorm.DB) error {
			return database.Exec(`
				DO $$ BEGIN
					IF EXISTS (
						SELECT 1 FROM information_schema.columns
						WHERE table_name = 'users' AND column_name = 'credits'
					) THEN
						ALTER TABLE users DROP COLUMN credits;
					END IF;
				END $$
			`).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Exec(
				"ALTER TABLE users ADD COLUMN IF NOT EXISTS credits INTEGER NOT NULL DEFAULT 0",
			).Error
		},
	}
}

// Step03UserBadges returns the migration that creates the user_badges table.
func Step03UserBadges() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_05_user_badges",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&inventorymodel.Badge{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&inventorymodel.Badge{})
		},
	}
}

// Step04UserEffects returns the migration that creates the user_effects table.
func Step04UserEffects() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_06_user_effects",
		Migrate: func(database *gorm.DB) error {
			return database.AutoMigrate(&inventorymodel.Effect{})
		},
		Rollback: func(database *gorm.DB) error {
			return database.Migrator().DropTable(&inventorymodel.Effect{})
		},
	}
}
