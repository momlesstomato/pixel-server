package application

import (
	"context"
	"fmt"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkfurniture "github.com/momlesstomato/pixel-sdk/events/furniture"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// Service defines furniture application use-cases.
type Service struct {
	// repository stores furniture persistence contract implementation.
	repository domain.Repository
	// fire stores optional plugin event dispatch behavior.
	fire func(sdk.Event)
}

// NewService creates one furniture service.
func NewService(repository domain.Repository) (*Service, error) {
	if repository == nil {
		return nil, fmt.Errorf("furniture repository is required")
	}
	return &Service{repository: repository}, nil
}

// SetEventFirer configures optional plugin event dispatch behavior.
func (service *Service) SetEventFirer(fire func(sdk.Event)) {
	service.fire = fire
}

// FindDefinitionByID resolves one item definition by identifier.
func (service *Service) FindDefinitionByID(ctx context.Context, id int) (domain.Definition, error) {
	if id <= 0 {
		return domain.Definition{}, fmt.Errorf("definition id must be positive")
	}
	return service.repository.FindDefinitionByID(ctx, id)
}

// ListDefinitions resolves all item definitions.
func (service *Service) ListDefinitions(ctx context.Context) ([]domain.Definition, error) {
	return service.repository.ListDefinitions(ctx)
}

// CreateDefinition persists one validated item definition.
func (service *Service) CreateDefinition(ctx context.Context, def domain.Definition) (domain.Definition, error) {
	if def.ItemName == "" {
		return domain.Definition{}, fmt.Errorf("item name is required")
	}
	if service.fire != nil {
		event := &sdkfurniture.DefinitionCreating{ItemName: def.ItemName}
		service.fire(event)
		if event.Cancelled() {
			return domain.Definition{}, fmt.Errorf("definition creation cancelled by plugin")
		}
	}
	result, err := service.repository.CreateDefinition(ctx, def)
	if err == nil && service.fire != nil {
		service.fire(&sdkfurniture.DefinitionCreated{DefinitionID: result.ID})
	}
	return result, err
}

// UpdateDefinition applies partial definition update.
func (service *Service) UpdateDefinition(ctx context.Context, id int, patch domain.DefinitionPatch) (domain.Definition, error) {
	if id <= 0 {
		return domain.Definition{}, fmt.Errorf("definition id must be positive")
	}
	return service.repository.UpdateDefinition(ctx, id, patch)
}

// DeleteDefinition removes one item definition by identifier.
func (service *Service) DeleteDefinition(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("definition id must be positive")
	}
	if service.fire != nil {
		event := &sdkfurniture.DefinitionDeleting{DefinitionID: id}
		service.fire(event)
		if event.Cancelled() {
			return fmt.Errorf("definition deletion cancelled by plugin")
		}
	}
	err := service.repository.DeleteDefinition(ctx, id)
	if err == nil && service.fire != nil {
		service.fire(&sdkfurniture.DefinitionDeleted{DefinitionID: id})
	}
	return err
}

// FindItemByID resolves one item instance by identifier.
func (service *Service) FindItemByID(ctx context.Context, id int) (domain.Item, error) {
	if id <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	return service.repository.FindItemByID(ctx, id)
}

// ListItemsByUserID resolves all items for one user.
func (service *Service) ListItemsByUserID(ctx context.Context, userID int) ([]domain.Item, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be positive")
	}
	return service.repository.ListItemsByUserID(ctx, userID)
}

// CreateItem persists one validated item instance.
func (service *Service) CreateItem(ctx context.Context, item domain.Item) (domain.Item, error) {
	if item.UserID <= 0 {
		return domain.Item{}, fmt.Errorf("user id must be positive")
	}
	if item.DefinitionID <= 0 {
		return domain.Item{}, fmt.Errorf("definition id must be positive")
	}
	return service.repository.CreateItem(ctx, item)
}

// TransferItem changes ownership of one item.
func (service *Service) TransferItem(ctx context.Context, itemID int, newUserID int) error {
	if itemID <= 0 {
		return fmt.Errorf("item id must be positive")
	}
	if newUserID <= 0 {
		return fmt.Errorf("new user id must be positive")
	}
	return service.repository.TransferItem(ctx, itemID, newUserID)
}

// DeleteItem removes one item by identifier.
func (service *Service) DeleteItem(ctx context.Context, id int) error {
	if id <= 0 {
		return fmt.Errorf("item id must be positive")
	}
	return service.repository.DeleteItem(ctx, id)
}
