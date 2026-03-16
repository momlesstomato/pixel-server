package domain

import (
	"errors"
)

// ErrUserNotFound defines missing user lookup behavior.
var ErrUserNotFound = errors.New("user not found")

// ErrRespectLimitReached defines daily respect limit behavior.
var ErrRespectLimitReached = errors.New("daily respect limit reached")

// ErrNameAlreadyTaken defines duplicate username change behavior.
var ErrNameAlreadyTaken = errors.New("username is already taken")

// ErrInvalidName defines invalid username format behavior.
var ErrInvalidName = errors.New("invalid username")

// ErrNameChangeNotAllowed defines rename-guard behavior for users.
var ErrNameChangeNotAllowed = errors.New("name change is not allowed")

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

// Profile defines one public user profile payload.
type Profile struct {
	// UserID stores stable user identifier.
	UserID int
	// Username stores current account username.
	Username string
	// Figure stores avatar figure string.
	Figure string
	// Motto stores profile motto value.
	Motto string
	// Registration stores account registration date string.
	Registration string
	// AchievementPoints stores profile achievement points.
	AchievementPoints int
	// FriendsCount stores total friends count.
	FriendsCount int
	// IsMyFriend stores whether viewer and target are friends.
	IsMyFriend bool
	// RequestSent stores whether viewer has a pending request to target.
	RequestSent bool
	// IsOnline stores current online marker.
	IsOnline bool
	// SecondsSinceLastVisit stores elapsed seconds since target last access.
	SecondsSinceLastVisit int
	// OpenProfileWindow stores client open profile flag.
	OpenProfileWindow bool
}
