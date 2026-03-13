package plugin

import (
	"reflect"
	"sort"
	"sync"
	"sync/atomic"

	sdk "github.com/momlesstomato/pixel-sdk"
	"go.uber.org/zap"
)

var nextHandlerID atomic.Uint64

// handlerEntry stores one registered event handler with its options.
type handlerEntry struct {
	id            uint64
	eventType     reflect.Type
	priority      sdk.Priority
	skipCancelled bool
	fn            func(sdk.Event)
	owner         string
}

// Dispatcher dispatches typed events to registered handlers.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[reflect.Type][]handlerEntry
	logger   *zap.Logger
}

// NewDispatcher creates a typed event dispatcher.
func NewDispatcher(logger *zap.Logger) *Dispatcher {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &Dispatcher{handlers: make(map[reflect.Type][]handlerEntry), logger: logger}
}

// Subscribe registers a handler for one concrete event type.
func (d *Dispatcher) Subscribe(owner string, handler any, opts ...sdk.HandlerOption) func() {
	fn, eventType := resolveHandler(handler)
	if fn == nil {
		panic("plugin: handler must be func(*ConcreteEventType)")
	}
	cfg := sdk.HandlerConfig{Priority: sdk.PriorityNormal}
	for _, opt := range opts {
		opt(&cfg)
	}
	entry := handlerEntry{id: nextHandlerID.Add(1), eventType: eventType, priority: cfg.Priority, skipCancelled: cfg.SkipCancelled, fn: fn, owner: owner}
	d.mu.Lock()
	d.handlers[eventType] = append(d.handlers[eventType], entry)
	sort.SliceStable(d.handlers[eventType], func(i, j int) bool {
		return d.handlers[eventType][i].priority < d.handlers[eventType][j].priority
	})
	d.mu.Unlock()
	id := entry.id
	return func() { d.unsubscribe(eventType, id) }
}

// Fire dispatches one event to all registered handlers by type.
func (d *Dispatcher) Fire(event sdk.Event) {
	t := reflect.TypeOf(event)
	d.mu.RLock()
	entries := d.handlers[t]
	d.mu.RUnlock()
	for i := range entries {
		e := &entries[i]
		if cancellable, ok := event.(sdk.Cancellable); ok && e.skipCancelled && cancellable.Cancelled() {
			if e.priority != sdk.PriorityMonitor {
				continue
			}
		}
		d.invoke(e, event)
	}
}

// RemoveByOwner removes all handlers registered by one plugin.
func (d *Dispatcher) RemoveByOwner(owner string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for t, entries := range d.handlers {
		filtered := entries[:0]
		for _, e := range entries {
			if e.owner != owner {
				filtered = append(filtered, e)
			}
		}
		d.handlers[t] = filtered
	}
}

// invoke executes one handler with panic recovery.
func (d *Dispatcher) invoke(e *handlerEntry, event sdk.Event) {
	defer func() {
		if r := recover(); r != nil {
			d.logger.Error("plugin handler panicked", zap.Any("panic", r), zap.String("event", e.eventType.String()), zap.String("owner", e.owner))
		}
	}()
	e.fn(event)
}

// unsubscribe removes one specific handler entry by ID.
func (d *Dispatcher) unsubscribe(t reflect.Type, id uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	entries := d.handlers[t]
	for i := range entries {
		if entries[i].id == id {
			d.handlers[t] = append(entries[:i], entries[i+1:]...)
			return
		}
	}
}

// resolveHandler extracts handler function and event type from an any parameter.
func resolveHandler(handler any) (func(sdk.Event), reflect.Type) {
	v := reflect.ValueOf(handler)
	t := v.Type()
	if t.Kind() != reflect.Func || t.NumIn() != 1 || t.NumOut() != 0 {
		return nil, nil
	}
	paramType := t.In(0)
	return func(event sdk.Event) { v.Call([]reflect.Value{reflect.ValueOf(event)}) }, paramType
}
