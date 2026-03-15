package domain

import "context"

// GroupPatch defines mutable group attribute payload.
type GroupPatch struct {
	// DisplayName stores optional display-name update.
	DisplayName *string
	// Priority stores optional priority update.
	Priority *int
	// ClubLevel stores optional club-level update.
	ClubLevel *int
	// SecurityLevel stores optional security-level update.
	SecurityLevel *int
	// IsAmbassador stores optional ambassador-flag update.
	IsAmbassador *bool
	// IsDefault stores optional default-flag update.
	IsDefault *bool
}

// Repository defines permission persistence behavior.
type Repository interface {
	// ListGroups returns all groups sorted by priority descending and id ascending.
	ListGroups(context.Context) ([]Group, error)
	// FindGroupByID resolves one group by identifier.
	FindGroupByID(context.Context, int) (Group, error)
	// FindGroupByName resolves one group by name.
	FindGroupByName(context.Context, string) (Group, error)
	// CreateGroup creates one new permission group.
	CreateGroup(context.Context, Group) (Group, error)
	// UpdateGroup updates mutable attributes of one group.
	UpdateGroup(context.Context, int, GroupPatch) (Group, error)
	// DeleteGroup deletes one group when constraints allow.
	DeleteGroup(context.Context, int) error
	// CountGroupUsers counts users assigned to one group.
	CountGroupUsers(context.Context, int) (int64, error)
	// FindDefaultGroup resolves the active default group.
	FindDefaultGroup(context.Context) (Group, error)
	// SwitchDefaultGroup marks one group as default and unmarks the previous default group.
	SwitchDefaultGroup(context.Context, int) error
	// ListGroupPermissions returns one group's granted permissions.
	ListGroupPermissions(context.Context, int) ([]string, error)
	// AddGroupPermissions adds permission grants to one group.
	AddGroupPermissions(context.Context, int, []string) error
	// RemoveGroupPermission removes one permission grant from one group.
	RemoveGroupPermission(context.Context, int, string) error
	// ListUserGroupIDs resolves assigned group identifiers for one user.
	ListUserGroupIDs(context.Context, int) ([]int, error)
	// ReplaceUserGroups replaces group assignments for one user.
	ReplaceUserGroups(context.Context, int, []int) error
}
