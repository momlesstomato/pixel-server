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

// hcShopPage defines the HC/subscription shop catalog page.
var hcShopPage = catalogmodel.Page{
	Caption: "HC Shop", Visible: true, Enabled: true, OrderNum: 10,
	IconImage: 10, PageLayout: "club_buy", ClubOnly: false,
}

var hcShopImages = []string{
	"catalogue/feature_cata_hort_HC_b.png",
	"catalogue/feature_cata_hort_HC_b.png",
}

// clubGiftsPage defines the HC monthly gifts catalog page.
var clubGiftsPage = catalogmodel.Page{
	Caption: "HC Gifts", Visible: true, Enabled: true, OrderNum: 11,
	IconImage: 11, PageLayout: "club_gifts", ClubOnly: true,
	Images: hcShopImages[0], Texts: "Choose your monthly HC gift",
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

// Step02HCShopPage returns seed step for the HC/subscription shop catalog page.
func Step02HCShopPage() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260405_S02_hc_shop_page",
		Migrate: func(database *gorm.DB) error {
			return database.FirstOrCreate(&hcShopPage, catalogmodel.Page{Caption: hcShopPage.Caption}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("caption = ?", hcShopPage.Caption).Delete(&catalogmodel.Page{}).Error
		},
	}
}

// Step03HCShopLocalizationBackfill returns a seed step that backfills the HC shop localization required by Nitro.
func Step03HCShopLocalizationBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_S04_hc_shop_localization_backfill",
		Migrate: func(database *gorm.DB) error {
			return database.Model(&catalogmodel.Page{}).
				Where("caption = ?", hcShopPage.Caption).
				Updates(map[string]any{"images": hcShopImages[0] + "|" + hcShopImages[1], "texts": ""}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Model(&catalogmodel.Page{}).
				Where("caption = ?", hcShopPage.Caption).
				Updates(map[string]any{"images": "", "texts": ""}).Error
		},
	}
}

// Step04ClubGiftsPage returns the seed step for the HC club_gifts page shell.
func Step04ClubGiftsPage() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260406_S05_club_gifts_page",
		Migrate: func(database *gorm.DB) error {
			return database.FirstOrCreate(&clubGiftsPage, catalogmodel.Page{Caption: clubGiftsPage.Caption}).Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Where("caption = ?", clubGiftsPage.Caption).Delete(&catalogmodel.Page{}).Error
		},
	}
}

// Step05HCShopVipBuyBackfill returns a seed step that aligns the HC shop shell with Nitro's vip_buy layout.
func Step05HCShopVipBuyBackfill() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "20260407_S08_hc_shop_vip_buy_backfill",
		Migrate: func(database *gorm.DB) error {
			return database.Model(&catalogmodel.Page{}).
				Where("caption = ?", hcShopPage.Caption).
				Update("page_layout", "vip_buy").Error
		},
		Rollback: func(database *gorm.DB) error {
			return database.Model(&catalogmodel.Page{}).
				Where("caption = ?", hcShopPage.Caption).
				Update("page_layout", "club_buy").Error
		},
	}
}
