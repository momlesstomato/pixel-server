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
}
