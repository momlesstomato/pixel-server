package httpapi

import (
	"context"

	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// Service defines navigator API behavior required by HTTP routes.
type Service interface {
	// ListCategories resolves all navigator categories.
	ListCategories(context.Context) ([]domain.Category, error)
	// FindCategoryByID resolves one navigator category by identifier.
	FindCategoryByID(context.Context, int) (domain.Category, error)
	// CreateCategory persists one validated navigator category.
	CreateCategory(context.Context, domain.Category) (domain.Category, error)
	// DeleteCategory removes one navigator category by identifier.
	DeleteCategory(context.Context, int) error
	// ListRooms resolves paginated rooms with optional filter.
	ListRooms(context.Context, domain.RoomFilter) ([]domain.Room, int, error)
	// FindRoomByID resolves one room by identifier.
	FindRoomByID(context.Context, int) (domain.Room, error)
	// CreateRoom persists one room row.
	CreateRoom(context.Context, domain.Room) (domain.Room, error)
	// DeleteRoom removes one room by identifier.
	DeleteRoom(context.Context, int) error
}
