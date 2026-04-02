package domain

import "errors"

// ErrCategoryNotFound defines missing navigator category lookup behavior.
var ErrCategoryNotFound = errors.New("navigator category not found")

// ErrRoomNotFound defines missing room lookup behavior.
var ErrRoomNotFound = errors.New("room not found")

// ErrSearchNotFound defines missing saved search lookup behavior.
var ErrSearchNotFound = errors.New("saved search not found")

// ErrFavouriteNotFound defines missing favourite lookup behavior.
var ErrFavouriteNotFound = errors.New("favourite not found")

// ErrFavouriteLimitReached defines per-user favourite limit behavior.
var ErrFavouriteLimitReached = errors.New("favourite limit reached")

// ErrFavouriteAlreadyExists defines duplicate favourite behavior.
var ErrFavouriteAlreadyExists = errors.New("favourite already exists")
