package room

import "context"

// Repository provides CRUD operations for Room entities.
type Repository interface {
	GetByID(ctx context.Context, id int32) (*Room, error)
	GetByOwner(ctx context.Context, ownerID int32) ([]*Room, error)
	Create(ctx context.Context, r *Room) error
	Update(ctx context.Context, r *Room) error
	Delete(ctx context.Context, id int32) error
	GetModel(ctx context.Context, name string) (*Model, error)
	Search(ctx context.Context, query string, limit int) ([]*Room, error)
	GetPopular(ctx context.Context, limit int) ([]*Room, error)
}
