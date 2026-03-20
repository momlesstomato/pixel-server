package domain

import "context"

const (
	// PermFloodBypass allows skipping message flood rate limiting.
	PermFloodBypass = "messenger.flood.bypass"
	// PermFriendsExtended grants the extended friend list capacity.
	PermFriendsExtended = "messenger.friends.extended"
	// PermFriendsUnlimited grants unlimited friend list capacity.
	PermFriendsUnlimited = "messenger.friends.unlimited"
)

// PermissionChecker defines permission resolution behavior required by the messenger domain.
type PermissionChecker interface {
	// HasPermission reports whether one user holds one dotted permission.
	HasPermission(ctx context.Context, userID int, permission string) (bool, error)
}
