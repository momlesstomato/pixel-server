package user

import "context"

// RoleRepository manages Role records and user–role assignments.
type RoleRepository interface {
	// GetByID loads a single role by its numeric ID.
	GetByID(ctx context.Context, id int32) (*Role, error)

	// GetAll returns every defined role.
	GetAll(ctx context.Context) ([]*Role, error)

	// GetForUser returns all roles currently assigned to a user.
	GetForUser(ctx context.Context, userID int32) ([]*Role, error)

	// Create persists a new role. If r.ID is zero an ID is assigned.
	Create(ctx context.Context, r *Role) error

	// Update persists changes to an existing role.
	Update(ctx context.Context, r *Role) error

	// Delete removes a role and all its user assignments.
	Delete(ctx context.Context, id int32) error

	// AssignRole grants a role to a user. Idempotent.
	AssignRole(ctx context.Context, userID, roleID int32) error

	// RevokeRole removes a role from a user. Idempotent.
	RevokeRole(ctx context.Context, userID, roleID int32) error

	// HasRole reports whether a specific role is currently assigned to a user.
	HasRole(ctx context.Context, userID, roleID int32) (bool, error)
}

// PermissionRepository manages the registry of named permissions.
// Permissions are referenced by name string; this repository provides the
// canonical definitions used in admin UIs and documentation.
type PermissionRepository interface {
	// GetAll returns every registered permission.
	GetAll(ctx context.Context) ([]*Permission, error)

	// GetByName loads a single permission by its name key.
	GetByName(ctx context.Context, name string) (*Permission, error)

	// Create registers a new permission. Overwrites if name already exists.
	Create(ctx context.Context, p *Permission) error

	// Delete removes a permission definition.
	// This does not revoke the permission from existing roles.
	Delete(ctx context.Context, name string) error
}
