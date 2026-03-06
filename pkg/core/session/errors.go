package session

import "errors"

// ErrEmptySessionID indicates session id is required.
var ErrEmptySessionID = errors.New("session id is required")

// ErrNilConnection indicates connection is required.
var ErrNilConnection = errors.New("connection is required")

// ErrSessionExists indicates session id is already registered.
var ErrSessionExists = errors.New("session already exists")

// ErrSessionNotFound indicates session id has no active connection.
var ErrSessionNotFound = errors.New("session not found")
