package config

import "errors"

// ErrEmptyAppEnv indicates that APP_ENV is required.
var ErrEmptyAppEnv = errors.New("app env is required")

// ErrEmptyRuntimeRole indicates that runtime role is required.
var ErrEmptyRuntimeRole = errors.New("runtime role is required")

// ErrEmptyRuntimeInstanceID indicates that runtime instance id is required.
var ErrEmptyRuntimeInstanceID = errors.New("runtime instance id is required")

// ErrInvalidRuntimeRole indicates that a runtime role value is unknown.
var ErrInvalidRuntimeRole = errors.New("invalid runtime role")
