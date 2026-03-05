package config

import "errors"

// ErrEmptyAppEnv indicates that APP_ENV is required.
var ErrEmptyAppEnv = errors.New("app env is required")
