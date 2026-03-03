package user

import "context"

// Repository provides CRUD operations for User entities.
type Repository interface {
	GetByID(ctx context.Context, id int32) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	GetSettings(ctx context.Context, userID int32) (*Settings, error)
	UpdateSettings(ctx context.Context, s *Settings) error
	SetOnline(ctx context.Context, userID int32, online bool) error
}

// WardrobeRepository manages saved outfit slots.
type WardrobeRepository interface {
	GetByUser(ctx context.Context, userID int32) ([]*WardrobeOutfit, error)
	Save(ctx context.Context, outfit *WardrobeOutfit) error
}

// BadgeRepository manages user badge collections.
type BadgeRepository interface {
	GetByUser(ctx context.Context, userID int32) ([]*Badge, error)
	GetEquipped(ctx context.Context, userID int32) ([]*Badge, error)
	Equip(ctx context.Context, userID int32, code string, slot int32) error
	Unequip(ctx context.Context, userID int32, slot int32) error
}

// IgnoreRepository manages user ignore lists.
type IgnoreRepository interface {
	GetByUser(ctx context.Context, userID int32) ([]*IgnoredUser, error)
	Add(ctx context.Context, userID int32, ignoredUserID int32) error
	Remove(ctx context.Context, userID int32, ignoredUserID int32) error
	IsIgnored(ctx context.Context, userID int32, targetID int32) (bool, error)
}

// SessionStore tracks online sessions.
type SessionStore interface {
	Set(ctx context.Context, userID int32, sessionID string) error
	Get(ctx context.Context, userID int32) (string, error)
	Delete(ctx context.Context, userID int32) error
	IsOnline(ctx context.Context, userID int32) (bool, error)
	Count(ctx context.Context) (int, error)
}
