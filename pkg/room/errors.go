package room

import "errors"

// Sentinel errors for the room domain.
var (
	// ErrNotFound is returned when a requested room entity does not exist.
	ErrNotFound = errors.New("room: not found")
)
