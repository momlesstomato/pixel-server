package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	navmodel "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	"gorm.io/gorm"
)

// defaultCategories defines essential bootstrap navigator categories
// matching Habbo standard flat categories presented in the navigator UI.
var defaultCategories = []navmodel.Category{
	{Caption: "No Category", Visible: true, OrderNum: 0, IconImage: 0, CategoryType: "public"},
	{Caption: "Trading", Visible: true, OrderNum: 1, IconImage: 1, CategoryType: "public"},
	{Caption: "Agencies & Fan Sites", Visible: true, OrderNum: 2, IconImage: 2, CategoryType: "public"},
	{Caption: "Games", Visible: true, OrderNum: 3, IconImage: 3, CategoryType: "public"},
	{Caption: "Help", Visible: true, OrderNum: 4, IconImage: 4, CategoryType: "public"},
	{Caption: "Life Style", Visible: true, OrderNum: 5, IconImage: 5, CategoryType: "public"},
	{Caption: "Party", Visible: true, OrderNum: 6, IconImage: 6, CategoryType: "public"},
	{Caption: "Role Playing", Visible: true, OrderNum: 7, IconImage: 7, CategoryType: "public"},
	{Caption: "Building & Decoration", Visible: true, OrderNum: 8, IconImage: 8, CategoryType: "public"},
	{Caption: "Chat & Discussion", Visible: true, OrderNum: 9, IconImage: 9, CategoryType: "public"},
	{Caption: "Personal Space", Visible: true, OrderNum: 10, IconImage: 10, CategoryType: "public"},
}

// Step01DefaultCategories returns seed step for essential navigator categories.
func Step01DefaultCategories() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260326_S01_nav_categories",
		Migrate: func(database *gorm.DB) error {
			for i := range defaultCategories {
				existing := navmodel.Category{}
				q := database.Where("caption = ?", defaultCategories[i].Caption).Limit(1).Find(&existing)
				if q.Error != nil {
					return q.Error
				}
				if q.RowsAffected > 0 {
					continue
				}
				if err := database.Create(&defaultCategories[i]).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			captions := make([]string, 0, len(defaultCategories))
			for _, c := range defaultCategories {
				captions = append(captions, c.Caption)
			}
			return database.Where("caption IN ?", captions).Delete(&navmodel.Category{}).Error
		},
	}
}
