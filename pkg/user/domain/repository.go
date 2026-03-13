package domain

import (
	"context"
	"time"
)

// ProfilePatch defines partial user identity update payload.
type ProfilePatch struct {
	// Figure stores optional avatar figure string.
	Figure *string
	// Gender stores optional avatar gender marker.
	Gender *string
	// Motto stores optional profile motto.
	Motto *string
	// HomeRoomID stores optional home room identifier.
	HomeRoomID *int
}

// RespectTargetType identifies respect target type.
type RespectTargetType int16

const (
	// RespectTargetUser identifies user-to-user respect.
	RespectTargetUser RespectTargetType = 0
	// RespectTargetPet identifies user-to-pet respect.
	RespectTargetPet RespectTargetType = 1
)

// Repository defines user persistence behavior.
type Repository interface {
	// Create persists one user row using the provided username.
	Create(context.Context, string) (User, error)
	// FindByID resolves one user by identifier.
	FindByID(context.Context, int) (User, error)
	// DeleteByID soft-deletes one user by identifier.
	DeleteByID(context.Context, int) error
	// UpdateProfile applies partial identity updates and returns updated user payload.
	UpdateProfile(context.Context, int, ProfilePatch) (User, error)
	// LoadSettings resolves user settings and lazily creates defaults when missing.
	LoadSettings(context.Context, int) (Settings, error)
	// SaveSettings applies partial settings update and returns updated settings payload.
	SaveSettings(context.Context, int, SettingsPatch) (Settings, error)
	// RecordRespect persists one respect event and returns updated respects received counter.
	RecordRespect(context.Context, int, int, RespectTargetType, time.Time) (int, error)
	// RemainingRespects returns remaining daily respects for one actor and target type.
	RemainingRespects(context.Context, int, RespectTargetType, time.Time) (int, error)
	// RecordLogin persists one successful login event and reports whether it is first login in UTC day.
	RecordLogin(context.Context, int, string, time.Time) (bool, error)
	// LoadWardrobe resolves saved wardrobe slots for one user.
	LoadWardrobe(context.Context, int) ([]WardrobeSlot, error)
	// SaveWardrobeSlot upserts one wardrobe slot payload for one user.
	SaveWardrobeSlot(context.Context, int, WardrobeSlot) error
	// ListIgnoredUsernames resolves ignored usernames for one user.
	ListIgnoredUsernames(context.Context, int) ([]string, error)
	// IgnoreUserByUsername stores one ignore relation by target username.
	IgnoreUserByUsername(context.Context, int, string) (int, error)
	// IgnoreUserByID stores one ignore relation by target user identifier.
	IgnoreUserByID(context.Context, int, int) error
	// UnignoreUserByUsername removes one ignore relation by target username.
	UnignoreUserByUsername(context.Context, int, string) (int, error)
	// LoadProfile resolves one partial public profile payload.
	LoadProfile(context.Context, int, bool) (Profile, error)
	// ListRespects resolves respect audit rows for one target user.
	ListRespects(context.Context, int, int, int) ([]RespectRecord, error)
	// IsUsernameAvailable checks whether one username is available.
	IsUsernameAvailable(context.Context, string, int) (bool, error)
	// ChangeUsername updates one user's username and rename flag.
	ChangeUsername(context.Context, int, string, bool) (User, error)
}
