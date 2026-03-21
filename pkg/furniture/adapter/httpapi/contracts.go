package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// Service defines furniture API behavior required by HTTP routes.
type Service interface {
	// FindDefinitionByID resolves one item definition by identifier.
	FindDefinitionByID(context.Context, int) (domain.Definition, error)
	// ListDefinitions resolves all item definitions.
	ListDefinitions(context.Context) ([]domain.Definition, error)
	// CreateDefinition persists one validated item definition.
	CreateDefinition(context.Context, domain.Definition) (domain.Definition, error)
	// UpdateDefinition applies partial definition update.
	UpdateDefinition(context.Context, int, domain.DefinitionPatch) (domain.Definition, error)
	// DeleteDefinition removes one item definition by identifier.
	DeleteDefinition(context.Context, int) error
	// FindItemByID resolves one item instance by identifier.
	FindItemByID(context.Context, int) (domain.Item, error)
	// ListItemsByUserID resolves all items for one user.
	ListItemsByUserID(context.Context, int) ([]domain.Item, error)
	// TransferItem changes ownership of one item.
	TransferItem(ctx context.Context, itemID int, newUserID int) error
}
