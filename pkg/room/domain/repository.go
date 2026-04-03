package domain

import (
	"context"
	"time"
)

// ModelRepository defines persistence operations for room model templates.
type ModelRepository interface {
	// FindModelBySlug resolves one room model by slug identifier.
	FindModelBySlug(ctx context.Context, slug string) (RoomModel, error)
	// ListModels returns all available room model templates.
	ListModels(ctx context.Context) ([]RoomModel, error)
}

// BanRepository defines persistence operations for room bans.
type BanRepository interface {
	// FindActiveBan resolves an active ban for one user in one room.
	FindActiveBan(ctx context.Context, roomID int, userID int) (*RoomBan, error)
	// CreateBan persists one room ban entry.
	CreateBan(ctx context.Context, ban RoomBan) (RoomBan, error)
	// DeleteBan removes one room ban by identifier.
	DeleteBan(ctx context.Context, id int) error
	// ListBansByRoom returns all active bans for one room.
	ListBansByRoom(ctx context.Context, roomID int) ([]RoomBan, error)
}

// RightsRepository defines persistence operations for room rights.
type RightsRepository interface {
	// HasRights reports whether one user holds rights in one room.
	HasRights(ctx context.Context, roomID int, userID int) (bool, error)
	// GrantRights adds rights for one user in one room.
	GrantRights(ctx context.Context, roomID int, userID int) error
	// RevokeRights removes rights for one user in one room.
	RevokeRights(ctx context.Context, roomID int, userID int) error
	// ListRightsByRoom returns all rights holders for one room.
	ListRightsByRoom(ctx context.Context, roomID int) ([]int, error)
	// RevokeAllRights removes all rights for one room.
	RevokeAllRights(ctx context.Context, roomID int) error
}

// RoomBan defines one room access ban entry.
type RoomBan struct {
	// ID stores the stable ban identifier.
	ID int
	// RoomID stores the room reference.
	RoomID int
	// UserID stores the banned user identifier.
	UserID int
	// ExpiresAt stores the ban expiry timestamp (nil = permanent).
	ExpiresAt *time.Time
	// CreatedAt stores the ban creation timestamp.
	CreatedAt time.Time
}

// RoomRight defines one room rights grant entry.
type RoomRight struct {
	// ID stores the stable rights identifier.
	ID int
	// RoomID stores the room reference.
	RoomID int
	// UserID stores the rights holder identifier.
	UserID int
}

// RoomRepository defines persistence operations for room aggregate data.
type RoomRepository interface {
	// FindByID resolves one full room record by stable identifier.
	FindByID(ctx context.Context, roomID int) (Room, error)
	// SaveSettings persists updated room settings for one room.
	SaveSettings(ctx context.Context, room Room) error
}
