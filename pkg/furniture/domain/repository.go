package domain

import "context"

// Repository defines furniture persistence behavior.
type Repository interface {
	// FindDefinitionByID resolves one item definition by identifier.
	FindDefinitionByID(context.Context, int) (Definition, error)
	// FindDefinitionByName resolves one item definition by internal name.
	FindDefinitionByName(context.Context, string) (Definition, error)
	// ListDefinitions resolves all item definition rows.
	ListDefinitions(context.Context) ([]Definition, error)
	// CreateDefinition persists one item definition row.
	CreateDefinition(context.Context, Definition) (Definition, error)
	// UpdateDefinition applies partial definition update.
	UpdateDefinition(context.Context, int, DefinitionPatch) (Definition, error)
	// DeleteDefinition removes one item definition by identifier.
	DeleteDefinition(context.Context, int) error
	// FindItemByID resolves one item instance by identifier.
	FindItemByID(context.Context, int) (Item, error)
	// ListItemsByUserID resolves all inventory items for one user.
	ListItemsByUserID(context.Context, int) ([]Item, error)
	// CreateItem persists one item instance.
	CreateItem(context.Context, Item) (Item, error)
	// DeleteItem removes one item instance by identifier.
	DeleteItem(context.Context, int) error
	// TransferItem changes item ownership atomically.
	TransferItem(ctx context.Context, itemID int, newUserID int) error
	// UpdateItemData updates the custom data payload for one item.
	UpdateItemData(ctx context.Context, itemID int, extraData string) error
	// UpdateItemInteractionData updates the hidden interaction payload for one item.
	UpdateItemInteractionData(ctx context.Context, itemID int, interactionData string) error
	// PlaceItem updates item room placement coordinates.
	PlaceItem(ctx context.Context, itemID int, roomID int, x int, y int, z float64, dir int) error
	// PlaceWallItem updates item wall placement for one room.
	PlaceWallItem(ctx context.Context, itemID int, roomID int, wallPosition string) error
	// UpdateItemDefinition transforms one item into another definition payload.
	UpdateItemDefinition(ctx context.Context, itemID int, definitionID int, extraData string, interactionData string) error
	// ListItemsByRoomID resolves all placed items in one room.
	ListItemsByRoomID(context.Context, int) ([]Item, error)
	// CountItemsByUserID returns item count for one user inventory.
	CountItemsByUserID(context.Context, int) (int, error)
}

// DefinitionPatch defines partial item definition update payload.
type DefinitionPatch struct {
	// PublicName stores optional display name update.
	PublicName *string
	// StackHeight stores optional stack height update.
	StackHeight *float64
	// CanStack stores optional stack flag update.
	CanStack *bool
	// CanLay stores optional lay flag update.
	CanLay *bool
	// AllowTrade stores optional trade flag update.
	AllowTrade *bool
	// AllowMarketplaceSell stores optional marketplace flag update.
	AllowMarketplaceSell *bool
	// AllowGift stores optional gift flag update.
	AllowGift *bool
	// InteractionType stores optional interaction type update.
	InteractionType *string
}
