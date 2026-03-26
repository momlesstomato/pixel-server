package store

import (
	"context"
	"errors"
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogmodel "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/model"
	"gorm.io/gorm"
)

// ListPages resolves all catalog page rows.
func (store *Store) ListPages(ctx context.Context) ([]domain.CatalogPage, error) {
	var rows []catalogmodel.Page
	if err := store.database.WithContext(ctx).Order("order_num ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.CatalogPage, len(rows))
	for i, row := range rows {
		result[i] = mapPage(row)
	}
	return result, nil
}

// FindPageByID resolves one catalog page by identifier.
func (store *Store) FindPageByID(ctx context.Context, id int) (domain.CatalogPage, error) {
	var row catalogmodel.Page
	if err := store.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.CatalogPage{}, domain.ErrPageNotFound
		}
		return domain.CatalogPage{}, err
	}
	return mapPage(row), nil
}

// CreatePage persists one catalog page row.
func (store *Store) CreatePage(ctx context.Context, page domain.CatalogPage) (domain.CatalogPage, error) {
	var parentID *uint
	if page.ParentID != nil {
		v := uint(*page.ParentID)
		parentID = &v
	}
	row := catalogmodel.Page{
		ParentID: parentID, Caption: page.Caption,
		IconImage: page.IconImage, PageLayout: page.PageLayout,
		Visible: page.Visible, Enabled: page.Enabled,
		MinPermission: page.MinPermission, ClubOnly: page.ClubOnly,
		OrderNum: page.OrderNum,
		Images: strings.Join(page.Images, "|"),
		Texts: strings.Join(page.Texts, "|"),
	}
	if err := store.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.CatalogPage{}, err
	}
	return mapPage(row), nil
}

// UpdatePage applies partial page update.
func (store *Store) UpdatePage(ctx context.Context, id int, patch domain.PagePatch) (domain.CatalogPage, error) {
	updates := map[string]any{}
	if patch.Caption != nil {
		updates["caption"] = *patch.Caption
	}
	if patch.Visible != nil {
		updates["visible"] = *patch.Visible
	}
	if patch.Enabled != nil {
		updates["enabled"] = *patch.Enabled
	}
	if patch.MinPermission != nil {
		updates["min_permission"] = *patch.MinPermission
	}
	if patch.OrderNum != nil {
		updates["order_num"] = *patch.OrderNum
	}
	if patch.PageLayout != nil {
		updates["page_layout"] = *patch.PageLayout
	}
	if len(updates) > 0 {
		result := store.database.WithContext(ctx).Model(&catalogmodel.Page{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return domain.CatalogPage{}, result.Error
		}
		if result.RowsAffected == 0 {
			return domain.CatalogPage{}, domain.ErrPageNotFound
		}
	}
	return store.FindPageByID(ctx, id)
}

// DeletePage removes one catalog page by identifier.
func (store *Store) DeletePage(ctx context.Context, id int) error {
	result := store.database.WithContext(ctx).Delete(&catalogmodel.Page{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrPageNotFound
	}
	return nil
}
