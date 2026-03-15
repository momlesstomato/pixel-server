package permission

import sdk "github.com/momlesstomato/pixel-sdk"

// PermissionChecked fires after one permission check is resolved.
type PermissionChecked struct {
	sdk.BaseEvent
	// UserID stores user identifier.
	UserID int
	// Permission stores checked permission string.
	Permission string
	// Granted stores check result.
	Granted bool
}
