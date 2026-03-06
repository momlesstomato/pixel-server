package local

import "sync"

// subscription is a local bus subscription handle.
type subscription struct {
	once sync.Once
	stop func() error
}

// newSubscription creates a new subscription handle.
func newSubscription(stop func() error) *subscription {
	return &subscription{stop: stop}
}

// Unsubscribe removes the subscription handler.
func (s *subscription) Unsubscribe() error {
	var err error
	s.once.Do(func() {
		err = s.stop()
	})
	return err
}
