package domain

import "context"

// Repository defines navigator persistence behavior.
type Repository interface {
	// ListCategories resolves all navigator category rows.
	ListCategories(context.Context) ([]Category, error)
	// FindCategoryByID resolves one navigator category by identifier.
	FindCategoryByID(context.Context, int) (Category, error)
	// CreateCategory persists one navigator category row.
	CreateCategory(context.Context, Category) (Category, error)
	// DeleteCategory removes one navigator category by identifier.
	DeleteCategory(context.Context, int) error
	// ListRooms resolves paginated rooms with optional filter.
	ListRooms(ctx context.Context, filter RoomFilter) ([]Room, int, error)
	// FindRoomByID resolves one room by identifier.
	FindRoomByID(context.Context, int) (Room, error)
	// CreateRoom persists one room row.
	CreateRoom(context.Context, Room) (Room, error)
	// UpdateRoom applies partial room update.
	UpdateRoom(context.Context, int, RoomPatch) (Room, error)
	// DeleteRoom removes one room by identifier.
	DeleteRoom(context.Context, int) error
	// ListSavedSearches resolves all saved searches for one user.
	ListSavedSearches(ctx context.Context, userID int) ([]SavedSearch, error)
	// CreateSavedSearch persists one saved search row.
	CreateSavedSearch(context.Context, SavedSearch) (SavedSearch, error)
	// DeleteSavedSearch removes one saved search by identifier.
	DeleteSavedSearch(ctx context.Context, id int) error
	// ListFavourites resolves all favourite room IDs for one user.
	ListFavourites(ctx context.Context, userID int) ([]Favourite, error)
	// AddFavourite creates one favourite entry for a user-room pair.
	AddFavourite(ctx context.Context, userID int, roomID int) error
	// RemoveFavourite deletes one favourite entry by user-room pair.
	RemoveFavourite(ctx context.Context, userID int, roomID int) error
	// CountFavourites returns favourite count for one user.
	CountFavourites(ctx context.Context, userID int) (int, error)
}

// RoomFilter defines navigator room search filter parameters.
type RoomFilter struct {
	// CategoryID stores optional category filter.
	CategoryID *int
	// SearchQuery stores optional text search filter.
	SearchQuery string
	// OwnerID stores optional owner filter.
	OwnerID *int
	// Offset stores pagination offset.
	Offset int
	// Limit stores pagination page size.
	Limit int
}

// RoomPatch defines partial room update payload.
type RoomPatch struct {
	// Name stores optional room name update.
	Name *string
	// Description stores optional room description update.
	Description *string
	// State stores optional access state update.
	State *string
	// CategoryID stores optional category update.
	CategoryID *int
	// MaxUsers stores optional capacity update.
	MaxUsers *int
}
