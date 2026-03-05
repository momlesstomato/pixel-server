package interfaces

import "errors"

// ErrNotFound indicates that a requested record does not exist.
var ErrNotFound = errors.New("record not found")
