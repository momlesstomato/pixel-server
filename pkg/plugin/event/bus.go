package event

import "sync"

type subscription struct {
	id      uint64
	handler Handler
}

type bus struct {
	mu      sync.RWMutex
	subs    map[Name][]subscription
	counter uint64
}

// NewBus creates an in-process synchronous event bus.
func NewBus() Bus {
	return &bus{
		subs: make(map[Name][]subscription),
	}
}

func (b *bus) Subscribe(name Name, handler Handler) CancelFunc {
	b.mu.Lock()
	b.counter++
	id := b.counter
	b.subs[name] = append(b.subs[name], subscription{id: id, handler: handler})
	b.mu.Unlock()

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		subs := b.subs[name]
		for i, s := range subs {
			if s.id == id {
				b.subs[name] = append(subs[:i], subs[i+1:]...)
				return
			}
		}
	}
}

func (b *bus) Publish(e *Event) {
	b.mu.RLock()
	subs := make([]subscription, len(b.subs[e.Name]))
	copy(subs, b.subs[e.Name])
	b.mu.RUnlock()

	for _, s := range subs {
		s.handler(e)
		if e.cancelled {
			return
		}
	}
}
