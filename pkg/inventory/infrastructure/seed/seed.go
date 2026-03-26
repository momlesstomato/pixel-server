package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	inventorymodel "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/model"
	"gorm.io/gorm"
)

// Step01CurrencyTypes seeds the three Habbo-standard activity-point currency
// type definitions. These IDs match the wire-protocol activityPointType field.
// Operators may insert additional rows into the currency_types table at any
// time; these three rows are bootstrap seeds only.
func Step01CurrencyTypes() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260324_inv_01_currency_types",
		Migrate: func(database *gorm.DB) error {
			types := []inventorymodel.CurrencyType{
				{ID: 0, Name: "duckets", DisplayName: "Duckets", Trackable: true, Enabled: true},
				{ID: 5, Name: "diamonds", DisplayName: "Diamonds", Trackable: true, Enabled: true},
				{ID: 105, Name: "seasonal", DisplayName: "Seasonal Points", Trackable: false, Enabled: true},
			}
			for _, ct := range types {
				existing := inventorymodel.CurrencyType{}
				query := database.Where("id = ?", ct.ID).Limit(1).Find(&existing)
				if query.Error != nil {
					return query.Error
				}
				if query.RowsAffected > 0 {
					continue
				}
				if err := database.Create(&ct).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("id IN ?", []int{0, 5, 105}).Delete(&inventorymodel.CurrencyType{}).Error
		},
	}
}
