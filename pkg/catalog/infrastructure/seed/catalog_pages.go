package seed

import (
	gormigrate "github.com/go-gormigrate/gormigrate/v2"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// defaultPages defines essential bootstrap catalog pages.
var defaultPages = []catalogmodel.Page{
	{Caption: "Frontpage", Visible: true, Enabled: true, OrderNum: 1, IconImage: 1, PageLayout: "frontpage4"},
	{Caption: "Furni", Visible: true, Enabled: true, OrderNum: 2, IconImage: 2, PageLayout: "default_3x3"},
	{Caption: "Walls", Visible: true, Enabled: true, OrderNum: 3, IconImage: 3, PageLayout: "default_3x3"},
	{Caption: "Pets", Visible: true, Enabled: true, OrderNum: 4, IconImage: 4, PageLayout: "pets"},
}

// Step01DefaultPages returns seed step for essential catalog pages.
func Step01DefaultPages() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260320_S01_catalog_pages",
		Migrate: func(database *gorm.DB) error {
			for i := range defaultPages {
				if err := database.FirstOrCreate(&defaultPages[i], catalogmodel.Page{Caption: defaultPages[i].Caption}).Error; err != nil {
					return err
				}
			}
			return nil
		},
		Rollback: func(database *gorm.DB) error {
			captions := make([]string, 0, len(defaultPages))
			for _, p := range defaultPages {
				captions = append(captions, p.Caption)
			}
			return database.Where("caption IN ?", captions).Delete(&catalogmodel.Page{}).Error
		},
	}
}
