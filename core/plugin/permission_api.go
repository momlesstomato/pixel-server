package plugin

import (
	"context"

	sdk "github.com/momlesstomato/pixel-sdk"
	sdkpermission "github.com/momlesstomato/pixel-sdk/events/permission"
)

// PermissionProvider defines permission and group resolution behavior for plugins.
type PermissionProvider interface {
	// HasPermission resolves whether one user has one permission.
	HasPermission(context.Context, int, string) (bool, error)
	// EffectiveGroup resolves one user's effective group snapshot.
	EffectiveGroup(context.Context, int) (sdk.GroupInfo, bool, error)
}

// pluginPermissionAPI provides plugin-facing permission resolution operations.
type pluginPermissionAPI struct {
	// provider stores permission resolution behavior.
	provider PermissionProvider
	// fire stores event dispatch behavior.
	fire func(sdk.Event)
	// emitChecks stores permission-check event emission behavior.
	emitChecks bool
}

// HasPermission reports whether one user has one permission.
func (api *pluginPermissionAPI) HasPermission(userID int, permission string) bool {
	if api.provider == nil || userID <= 0 {
		return false
	}
	granted, err := api.provider.HasPermission(context.Background(), userID, permission)
	if err != nil {
		return false
	}
	if api.emitChecks && api.fire != nil {
		api.fire(&sdkpermission.PermissionChecked{UserID: userID, Permission: permission, Granted: granted})
	}
	return granted
}

// GetGroup resolves the effective group for one user.
func (api *pluginPermissionAPI) GetGroup(userID int) (sdk.GroupInfo, bool) {
	if api.provider == nil || userID <= 0 {
		return sdk.GroupInfo{}, false
	}
	group, ok, err := api.provider.EffectiveGroup(context.Background(), userID)
	if err != nil {
		return sdk.GroupInfo{}, false
	}
	return group, ok
}
