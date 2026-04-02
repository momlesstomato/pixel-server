package store

import (
	"context"
	"errors"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
	model "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/model"
	"gorm.io/gorm"
)

// ListCategories resolves all navigator category rows.
func (s *Store) ListCategories(ctx context.Context) ([]domain.Category, error) {
	var rows []model.Category
	if err := s.database.WithContext(ctx).Order("order_num ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]domain.Category, len(rows))
	for i, row := range rows {
		result[i] = mapCategory(row)
	}
	return result, nil
}

// FindCategoryByID resolves one navigator category by identifier.
func (s *Store) FindCategoryByID(ctx context.Context, id int) (domain.Category, error) {
	var row model.Category
	if err := s.database.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Category{}, domain.ErrCategoryNotFound
		}
		return domain.Category{}, err
	}
	return mapCategory(row), nil
}

// CreateCategory persists one navigator category row.
func (s *Store) CreateCategory(ctx context.Context, cat domain.Category) (domain.Category, error) {
	row := model.Category{
		Caption: cat.Caption, Visible: cat.Visible, OrderNum: cat.OrderNum,
		IconImage: cat.IconImage, CategoryType: cat.CategoryType,
	}
	if err := s.database.WithContext(ctx).Create(&row).Error; err != nil {
		return domain.Category{}, err
	}
	return mapCategory(row), nil
}

// DeleteCategory removes one navigator category by identifier.
func (s *Store) DeleteCategory(ctx context.Context, id int) error {
	result := s.database.WithContext(ctx).Delete(&model.Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrCategoryNotFound
	}
	return nil
}
