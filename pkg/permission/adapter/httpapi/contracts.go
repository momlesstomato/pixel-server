package httpapi

import (
	"context"

	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// Service defines permission API behavior required by HTTP routes.
type Service interface {
	// ListGroups returns all groups with permissions.
	ListGroups(context.Context) ([]permissionapplication.GroupDetails, error)
	// GetGroup resolves one group by identifier.
	GetGroup(context.Context, int) (permissionapplication.GroupDetails, error)
	// CreateGroup creates one group from input payload.
	CreateGroup(context.Context, permissionapplication.CreateGroupInput) (permissionapplication.GroupDetails, error)
	// UpdateGroup updates one group from patch payload.
	UpdateGroup(context.Context, int, permissiondomain.GroupPatch) (permissionapplication.GroupDetails, error)
	// DeleteGroup deletes one group.
	DeleteGroup(context.Context, int) error
	// AddPermissions grants permissions to one group.
	AddPermissions(context.Context, int, []string) (permissionapplication.GroupDetails, error)
	// RemovePermission revokes one permission from one group.
	RemovePermission(context.Context, int, string) (permissionapplication.GroupDetails, error)
	// ReplaceUserGroups replaces group assignments for one user.
	ReplaceUserGroups(context.Context, int, []int) (permissiondomain.Access, error)
}
