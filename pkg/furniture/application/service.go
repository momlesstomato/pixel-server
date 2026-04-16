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
	if err := validateDefinitionCreate(def); err != nil {
		return domain.Definition{}, err
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

// PlaceFloorItem moves an owned inventory item into a room at the given coordinates.
// It validates ownership and returns the updated item after placement.
func (service *Service) PlaceFloorItem(ctx context.Context, itemID int, userID int, roomID int, x int, y int, dir int) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	item, err := service.repository.FindItemByID(ctx, itemID)
	if err != nil {
		return domain.Item{}, err
	}
	if item.RoomID == 0 && item.UserID != userID {
		return domain.Item{}, domain.ErrItemNotOwned
	}
	if item.RoomID != 0 && item.RoomID != roomID {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if err := service.repository.PlaceItem(ctx, itemID, roomID, x, y, 0, dir); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// ListRoomItems resolves all placed items in one room.
func (service *Service) ListRoomItems(ctx context.Context, roomID int) ([]domain.Item, error) {
	if roomID <= 0 {
		return nil, fmt.Errorf("room id must be positive")
	}
	return service.repository.ListItemsByRoomID(ctx, roomID)
}

// PickupItem removes an owned placed item from a room and returns it to inventory.
// It validates ownership and clears all placement coordinates.
func (service *Service) PickupItem(ctx context.Context, itemID int, userID int) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	item, err := service.repository.FindItemByID(ctx, itemID)
	if err != nil {
		return domain.Item{}, err
	}
	if item.RoomID == 0 && item.UserID != userID {
		return domain.Item{}, domain.ErrItemNotOwned
	}
	if err := service.repository.PlaceItem(ctx, itemID, 0, 0, 0, 0, 0); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// PlaceWallItem moves an owned inventory item onto a room wall anchor.
func (service *Service) PlaceWallItem(ctx context.Context, itemID int, userID int, roomID int, wallPosition string) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	item, err := service.repository.FindItemByID(ctx, itemID)
	if err != nil {
		return domain.Item{}, err
	}
	if item.RoomID == 0 && item.UserID != userID {
		return domain.Item{}, domain.ErrItemNotOwned
	}
	if item.RoomID != 0 && item.RoomID != roomID {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if err := service.repository.PlaceWallItem(ctx, itemID, roomID, wallPosition); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// UpdateItemData persists a visible item data payload and returns the updated item.
func (service *Service) UpdateItemData(ctx context.Context, itemID int, extraData string) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	if err := service.repository.UpdateItemData(ctx, itemID, extraData); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// UpdateItemInteractionData persists hidden interaction metadata and returns the updated item.
func (service *Service) UpdateItemInteractionData(ctx context.Context, itemID int, interactionData string) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	if err := service.repository.UpdateItemInteractionData(ctx, itemID, interactionData); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// TransformItem updates the definition and payload of one existing item.
func (service *Service) TransformItem(ctx context.Context, itemID int, definitionID int, extraData string, interactionData string) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	if definitionID <= 0 {
		return domain.Item{}, fmt.Errorf("definition id must be positive")
	}
	if err := service.repository.UpdateItemDefinition(ctx, itemID, definitionID, extraData, interactionData); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}

// MovePlacedItem updates one already-placed floor item including its height offset.
func (service *Service) MovePlacedItem(ctx context.Context, itemID int, roomID int, x int, y int, z float64, dir int) (domain.Item, error) {
	if itemID <= 0 {
		return domain.Item{}, fmt.Errorf("item id must be positive")
	}
	item, err := service.repository.FindItemByID(ctx, itemID)
	if err != nil {
		return domain.Item{}, err
	}
	if item.RoomID != roomID {
		return domain.Item{}, domain.ErrItemNotFound
	}
	if err := service.repository.PlaceItem(ctx, itemID, roomID, x, y, z, dir); err != nil {
		return domain.Item{}, err
	}
	return service.repository.FindItemByID(ctx, itemID)
}
