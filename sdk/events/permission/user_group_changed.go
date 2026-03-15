package permission

import sdk "github.com/momlesstomato/pixel-sdk"

// UserGroupChanged fires before user group assignments are updated.
type UserGroupChanged struct {
	sdk.BaseCancellable
	// UserID stores user identifier.
	UserID int
	// OldGroupID stores previous effective group identifier.
	OldGroupID int
	// NewGroupID stores next effective group identifier.
	NewGroupID int
	// OldGroupIDs stores previous assigned group identifiers.
	OldGroupIDs []int
	// NewGroupIDs stores next assigned group identifiers.
	NewGroupIDs []int
}
