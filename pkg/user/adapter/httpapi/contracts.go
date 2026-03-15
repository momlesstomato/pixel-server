package httpapi

import (
	"context"
	"time"

	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// Service defines user API behavior required by HTTP routes.
type Service interface {
	// FindByID resolves one user by identifier.
	FindByID(context.Context, int) (domain.User, error)
	// UpdateProfile validates and applies user identity updates.
	UpdateProfile(context.Context, int, domain.ProfilePatch) (domain.User, error)
	// LoadSettings resolves one user settings payload.
	LoadSettings(context.Context, int) (domain.Settings, error)
	// SaveSettings validates and applies one partial settings payload.
	SaveSettings(context.Context, int, domain.SettingsPatch) (domain.Settings, error)
	// RecordUserRespect stores one user-to-user respect event.
	RecordUserRespect(context.Context, int, int, time.Time) (userapplication.RespectResult, error)
	// LoadWardrobe resolves saved wardrobe slots for one user.
	LoadWardrobe(context.Context, int) ([]domain.WardrobeSlot, error)
	// ListRespects resolves respect audit rows for one target user.
	ListRespects(context.Context, int, int, int) ([]domain.RespectRecord, error)
	// ForceChangeName applies one administrative user rename operation.
	ForceChangeName(context.Context, int, string) (domain.NameResult, error)
	// ListIgnoredUsers resolves ignored user entries for one user.
	ListIgnoredUsers(context.Context, int) ([]domain.IgnoreEntry, error)
	// AdminIgnoreUser stores one admin-initiated ignore relation.
	AdminIgnoreUser(context.Context, int, int) error
	// AdminUnignoreUser removes one admin-initiated ignore relation.
	AdminUnignoreUser(context.Context, int, int) error
}
