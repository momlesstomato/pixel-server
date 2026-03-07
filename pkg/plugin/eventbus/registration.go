package eventbus

import (
	"sync"

	"pixelsv/pkg/plugin"
)

// registration tracks one event handler subscription.
type registration struct {
	// bus is the owning event bus instance.
	bus *Bus
	// event is the subscribed event name.
	event string
	// id is the registration identifier.
	id uint64
	// once guarantees one-time unsubscribe behavior.
	once sync.Once
}

// Unsubscribe removes the associated handler from the event bus.
func (r *registration) Unsubscribe() {
	if r == nil {
		return
	}
	r.once.Do(func() {
		if r.bus != nil {
			r.bus.removeHandler(r.event, r.id)
		}
	})
}

// noopRegistration is returned for ignored subscriptions.
type noopRegistration struct{}

// Unsubscribe is a no-op for ignored subscriptions.
func (n noopRegistration) Unsubscribe() {
	_ = n
}

var _ plugin.Registration = (*registration)(nil)
var _ plugin.Registration = noopRegistration{}
