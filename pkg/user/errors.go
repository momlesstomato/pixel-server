package user

import "errors"

// Sentinel errors for the user domain.
var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("user: not found")

	// ErrAlreadyExists is returned when creating a duplicate entity.
	ErrAlreadyExists = errors.New("user: already exists")
)
