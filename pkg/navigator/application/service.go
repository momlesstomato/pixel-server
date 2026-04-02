package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// MaxFavourites defines the per-user favourite room limit.
const MaxFavourites = 30

// MaxRoomsPerm defines the maximum rooms a user can own.
const MaxRoomsPerm = 25

// Service defines navigator application use-cases.
type Service struct {
	// repository stores navigator persistence contract implementation.
	repository domain.Repository
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one navigator service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("navigator repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// ListCategories resolves all navigator categories.
func (service *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return service.repository.ListCategories(ctx)
}

// FindCategoryByID resolves one navigator category by identifier.
func (service *Service) FindCategoryByID(ctx context.Context, id int) (domain.Category, error) {
	if id <= 0 {
		return domain.Category{}, fmt.Errorf("category id must be positive")
	}
	return service.repository.FindCategoryByID(ctx, id)
}

// CreateCategory persists one validated navigator category.
func (service *Service) CreateCategory(ctx context.Context, cat domain.Category) (domain.Category, error) {
	if cat.Caption == "" {
		return domain.Category{}, fmt.Errorf("category caption is required")
	}
	return service.repository.CreateCategory(ctx, cat)
}

// DeleteCategory removes one navigator category by identifier.
func (service *Service) DeleteCategory(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("category id must be positive")
	}
	return service.repository.DeleteCategory(ctx, id)
}
