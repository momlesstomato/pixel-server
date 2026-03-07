package eventbus

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"pixelsv/pkg/plugin"
)

// PanicHandler handles recovered panics from event handlers.
type PanicHandler func(event string, recovered any)

// Bus is an in-process event dispatch implementation for plugin hooks.
type Bus struct {
	// mu protects handler maps and registration mutations.
	mu sync.RWMutex
	// handlers indexes handlers by event name and registration identifier.
	handlers map[string]map[uint64]plugin.EventHandler
	// nextID provides monotonic registration identifiers.
	nextID atomic.Uint64
	// onPanic receives recovered panic details.
	onPanic PanicHandler
}

// New creates an event bus with no panic callback.
func New() *Bus {
	return NewWithPanicHandler(nil)
}

// NewWithPanicHandler creates an event bus with custom panic callback.
func NewWithPanicHandler(handler PanicHandler) *Bus {
	return &Bus{handlers: make(map[string]map[uint64]plugin.EventHandler), onPanic: handler}
}

// On registers one event handler and returns a registration handle.
func (b *Bus) On(event string, handler plugin.EventHandler) plugin.Registration {
	if b == nil || event == "" || handler == nil {
		return noopRegistration{}
	}
	id := b.nextID.Add(1)
	b.mu.Lock()
	if _, ok := b.handlers[event]; !ok {
		b.handlers[event] = make(map[uint64]plugin.EventHandler)
	}
	b.handlers[event][id] = handler
	b.mu.Unlock()
	return &registration{bus: b, event: event, id: id}
}

// Emit dispatches one event to every handler registered for its name.
func (b *Bus) Emit(event *plugin.Event) error {
	if b == nil {
		return nil
	}
	if event == nil {
		return ErrNilEvent
	}
	if event.Name == "" {
		return ErrMissingEventName
	}
	handlers := b.snapshot(event.Name)
	if len(handlers) == 0 {
		return nil
	}
	var errs []error
	for _, h := range handlers {
		if err := b.call(event.Name, h.id, h.handler, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// handlerEntry is an immutable handler snapshot used during emit.
type handlerEntry struct {
	// id is the registration identifier.
	id uint64
	// handler is the event callback.
	handler plugin.EventHandler
}

// snapshot copies all handlers for one event into a stable slice.
func (b *Bus) snapshot(event string) []handlerEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	set := b.handlers[event]
	if len(set) == 0 {
		return nil
	}
	result := make([]handlerEntry, 0, len(set))
	for id, handler := range set {
		result = append(result, handlerEntry{id: id, handler: handler})
	}
	return result
}

// call executes one event handler with panic recovery and deregistration.
func (b *Bus) call(event string, id uint64, handler plugin.EventHandler, payload *plugin.Event) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		if b.onPanic != nil {
			b.onPanic(event, recovered)
		}
		b.removeHandler(event, id)
		err = fmt.Errorf("plugin event handler panic: %v", recovered)
	}()
	return handler(payload)
}

// removeHandler deletes one handler registration from the event map.
func (b *Bus) removeHandler(event string, id uint64) {
	if b == nil || event == "" || id == 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if set, ok := b.handlers[event]; ok {
		delete(set, id)
		if len(set) == 0 {
			delete(b.handlers, event)
		}
	}
}

var _ plugin.EventBus = (*Bus)(nil)
