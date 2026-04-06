package migration

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// userCreditsRow is a minimal table proxy used only to check or drop the legacy credits column.
type userCreditsRow struct {
	// Credits stores the legacy credits column to be removed.
	Credits int `gorm:"column:credits"`
}

// TableName returns the users table name for this proxy struct.
func (userCreditsRow) TableName() string { return "users" }

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
			if !database.Migrator().HasColumn(&userCreditsRow{}, "credits") {
				return nil
			}
			return database.Migrator().DropColumn(&userCreditsRow{}, "credits")
		},
		Rollback: func(database *gorm.DB) error {
			if database.Migrator().HasColumn(&userCreditsRow{}, "credits") {
				return nil
			}
			return database.Migrator().AddColumn(&userCreditsRow{}, "credits")
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
