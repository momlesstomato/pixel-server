package domain

import "context"

// PermStaffPick defines the permission required to staff-pick a navigator room.
const PermStaffPick = "navigator.rooms.staff_pick"

// PermissionChecker defines permission resolution behavior required by the navigator domain.
type PermissionChecker interface {
	// HasPermission reports whether one user holds one named permission.
	HasPermission(ctx context.Context, userID int, perm string) (bool, error)
}
