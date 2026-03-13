package domain

import (
	"context"
	"errors"
	"time"
)

// ErrUserNotFound defines missing user lookup behavior.
var ErrUserNotFound = errors.New("user not found")

// User defines one user aggregate identity payload.
type User struct {
	// ID stores stable user identifier.
	ID int
	// Username stores user visible name.
	Username string
}

// Repository defines user persistence behavior.
type Repository interface {
	// Create persists one user row using the provided username.
	Create(context.Context, string) (User, error)
	// FindByID resolves one user by identifier.
	FindByID(context.Context, int) (User, error)
	// DeleteByID soft-deletes one user by identifier.
	DeleteByID(context.Context, int) error
	// RecordLogin persists one successful login event and reports whether it is first login in UTC day.
	RecordLogin(context.Context, int, string, time.Time) (bool, error)
}
