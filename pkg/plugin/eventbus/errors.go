package eventbus

import "errors"

var (
	// ErrNilEvent is returned when Emit receives a nil event pointer.
	ErrNilEvent = errors.New("plugin event is nil")
	// ErrMissingEventName is returned when Emit receives an event without name.
	ErrMissingEventName = errors.New("plugin event name is required")
)
