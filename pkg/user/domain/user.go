package domain

import (
	"errors"
)

// ErrUserNotFound defines missing user lookup behavior.
var ErrUserNotFound = errors.New("user not found")

// ErrRespectLimitReached defines daily respect limit behavior.
var ErrRespectLimitReached = errors.New("daily respect limit reached")

// User defines one user aggregate identity payload.
type User struct {
	// ID stores stable user identifier.
	ID int
	// Username stores user visible name.
	Username string
	// Figure stores avatar figure string.
	Figure string
	// Gender stores avatar gender marker.
	Gender string
	// Motto stores profile motto.
	Motto string
	// RealName stores profile real-name value.
	RealName string
	// RespectsReceived stores total received respects.
	RespectsReceived int
	// HomeRoomID stores configured home room identifier.
	HomeRoomID int
	// CanChangeName stores rename capability marker.
	CanChangeName bool
	// NoobnessLevel stores account age tier marker.
	NoobnessLevel int
	// SafetyLocked stores account safety lock marker.
	SafetyLocked bool
	// GroupID stores permission group identifier.
	GroupID int
}
