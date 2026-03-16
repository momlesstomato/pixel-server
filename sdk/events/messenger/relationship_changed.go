package messenger

import sdk "github.com/momlesstomato/pixel-sdk"

// RelationshipChanged fires before a friendship relationship type is updated.
type RelationshipChanged struct {
	sdk.BaseCancellable
	// UserID stores the user setting the relationship.
	UserID int
	// FriendUserID stores the friend whose relationship label changes.
	FriendUserID int
	// OldType stores the previous relationship type value.
	OldType int
	// NewType stores the new relationship type value.
	NewType int
}
