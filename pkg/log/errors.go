package log

import "errors"

// ErrInvalidFormat indicates that logging format is not supported.
var ErrInvalidFormat = errors.New("invalid log format")

// ErrEmptyLevel indicates that log level is required.
var ErrEmptyLevel = errors.New("log level is required")
