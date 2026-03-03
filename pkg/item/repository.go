package item

import "context"

// Repository provides CRUD operations for Item entities.
type Repository interface {
	GetByID(ctx context.Context, id int32) (*Item, error)
	GetByRoom(ctx context.Context, roomID int32) ([]*Item, error)
	GetByUser(ctx context.Context, userID int32) ([]*Item, error)
	Create(ctx context.Context, it *Item) error
	Update(ctx context.Context, it *Item) error
	Delete(ctx context.Context, id int32) error
	GetFurniture(ctx context.Context, baseItem int32) (*Furniture, error)
}
