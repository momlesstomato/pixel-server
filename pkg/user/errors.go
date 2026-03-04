package user

import "errors"

// Sentinel errors for the user domain.
var (
	// ErrNotFound is returned when a requested user entity does not exist.
	ErrNotFound = errors.New("user: not found")

	// ErrAlreadyExists is returned when creating a duplicate entity.
	ErrAlreadyExists = errors.New("user: already exists")

	// ErrRoleNotFound is returned when a requested role does not exist.
	ErrRoleNotFound = errors.New("user: role not found")

	// ErrPermissionNotFound is returned when a requested permission does not exist.
	ErrPermissionNotFound = errors.New("user: permission not found")
)
