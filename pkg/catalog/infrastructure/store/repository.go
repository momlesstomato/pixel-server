package store

import (
	"fmt"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// Store persists catalog data using PostgreSQL via GORM.
type Store struct {
	// database stores the ORM client reference.
	database *gorm.DB
}

// NewRepository creates one PostgreSQL catalog repository.
func NewRepository(database *gorm.DB) (*Store, error) {
	if database == nil {
		return nil, fmt.Errorf("postgres database is required")
	}
	return &Store{database: database}, nil
}

// compile-time interface assertion.
var _ domain.Repository = (*Store)(nil)

// mapPage converts one GORM model into domain catalog page.
func mapPage(row catalogmodel.Page) domain.CatalogPage {
	var images, texts []string
	if row.Images != "" {
		images = strings.Split(row.Images, ",")
	}
	if row.Texts != "" {
		texts = strings.Split(row.Texts, ",")
	}
	var parentID *int
	if row.ParentID != nil {
		v := int(*row.ParentID)
		parentID = &v
	}
	return domain.CatalogPage{
		ID: int(row.ID), ParentID: parentID, Caption: row.Caption,
		IconImage: row.IconImage, PageLayout: row.PageLayout,
		Visible: row.Visible, Enabled: row.Enabled,
		MinRank: row.MinRank, ClubOnly: row.ClubOnly,
		OrderNum: row.OrderNum, Images: images, Texts: texts,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// mapOffer converts one GORM model into domain catalog offer.
func mapOffer(row catalogmodel.Offer) domain.CatalogOffer {
	return domain.CatalogOffer{
		ID: int(row.ID), PageID: int(row.PageID),
		ItemDefinitionID: int(row.ItemDefinitionID),
		CatalogName: row.CatalogName,
		CostPrimary: row.CostPrimary, CostPrimaryType: row.CostPrimaryType,
		CostSecondary: row.CostSecondary, CostSecondaryType: row.CostSecondaryType,
		Amount: row.Amount, LimitedTotal: row.LimitedTotal,
		LimitedSells: row.LimitedSells, OfferActive: row.OfferActive,
		ExtraData: row.ExtraData, BadgeID: row.BadgeID,
		ClubOnly: row.ClubOnly, OrderNum: row.OrderNum,
	}
}

// mapVoucher converts one GORM model into domain voucher.
func mapVoucher(row catalogmodel.Voucher) domain.Voucher {
	return domain.Voucher{
		ID: int(row.ID), Code: row.Code,
		RewardType: row.RewardType, RewardCurrencyType: row.RewardCurrencyType,
		RewardData: row.RewardData,
		MaxUses: row.MaxUses, CurrentUses: row.CurrentUses,
		Enabled: row.Enabled, CreatedAt: row.CreatedAt,
	}
}
