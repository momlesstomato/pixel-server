package interceptor

import (
	"sync"
	"sync/atomic"

	"pixelsv/pkg/plugin"
)

// PanicHandler handles recovered panics from packet hooks.
type PanicHandler func(headerID uint16, recovered any)

// Interceptor implements plugin packet hook registration and execution.
type Interceptor struct {
	// mu protects hook registrations and snapshots.
	mu sync.RWMutex
	// before stores pre-hooks keyed by packet header.
	before map[uint16]map[uint64]plugin.PacketHook
	// after stores post-hooks keyed by packet header.
	after map[uint16]map[uint64]plugin.PacketHook
	// beforeAll stores pre-hooks for every header.
	beforeAll map[uint64]plugin.PacketHook
	// afterAll stores post-hooks for every header.
	afterAll map[uint64]plugin.PacketHook
	// nextID provides monotonic registration identifiers.
	nextID atomic.Uint64
	// onPanic receives panic details from hook execution.
	onPanic PanicHandler
}

// New creates an interceptor without panic callback.
func New() *Interceptor {
	return NewWithPanicHandler(nil)
}

// NewWithPanicHandler creates an interceptor with panic callback.
func NewWithPanicHandler(handler PanicHandler) *Interceptor {
	return &Interceptor{
		before:    make(map[uint16]map[uint64]plugin.PacketHook),
		after:     make(map[uint16]map[uint64]plugin.PacketHook),
		beforeAll: make(map[uint64]plugin.PacketHook),
		afterAll:  make(map[uint64]plugin.PacketHook),
		onPanic:   handler,
	}
}

// Before registers one pre-handler for one packet header.
func (i *Interceptor) Before(headerID uint16, hook plugin.PacketHook) plugin.Registration {
	return i.add(headerID, hook, true, false)
}

// After registers one post-handler for one packet header.
func (i *Interceptor) After(headerID uint16, hook plugin.PacketHook) plugin.Registration {
	return i.add(headerID, hook, false, false)
}

// BeforeAll registers one pre-handler for every packet.
func (i *Interceptor) BeforeAll(hook plugin.PacketHook) plugin.Registration {
	return i.add(0, hook, true, true)
}

// AfterAll registers one post-handler for every packet.
func (i *Interceptor) AfterAll(hook plugin.PacketHook) plugin.Registration {
	return i.add(0, hook, false, true)
}

// RunBefore executes pre-hooks and returns false when cancelled.
func (i *Interceptor) RunBefore(ctx plugin.PacketContext) bool {
	for _, entry := range i.snapshot(ctx.HeaderID, true) {
		if !i.call(ctx, entry) {
			return false
		}
	}
	return true
}

// RunAfter executes post-hooks after packet handling.
func (i *Interceptor) RunAfter(ctx plugin.PacketContext) {
	for _, entry := range i.snapshot(ctx.HeaderID, false) {
		i.call(ctx, entry)
	}
}

// hookEntry is an immutable hook snapshot used during execution.
type hookEntry struct {
	// id is the hook registration identifier.
	id uint64
	// all reports whether hook applies to all headers.
	all bool
	// pre reports whether hook is pre-handler.
	pre bool
	// hook is the callback function.
	hook plugin.PacketHook
}

// snapshot copies active hooks in execution order.
func (i *Interceptor) snapshot(headerID uint16, pre bool) []hookEntry {
	i.mu.RLock()
	defer i.mu.RUnlock()
	result := make([]hookEntry, 0)
	if pre {
		for id, hook := range i.beforeAll {
			result = append(result, hookEntry{id: id, all: true, pre: true, hook: hook})
		}
		for id, hook := range i.before[headerID] {
			result = append(result, hookEntry{id: id, all: false, pre: true, hook: hook})
		}
		return result
	}
	for id, hook := range i.afterAll {
		result = append(result, hookEntry{id: id, all: true, pre: false, hook: hook})
	}
	for id, hook := range i.after[headerID] {
		result = append(result, hookEntry{id: id, all: false, pre: false, hook: hook})
	}
	return result
}

// call executes one hook with panic recovery and hot removal.
func (i *Interceptor) call(ctx plugin.PacketContext, entry hookEntry) (allowed bool) {
	allowed = true
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		if i.onPanic != nil {
			i.onPanic(ctx.HeaderID, recovered)
		}
		i.remove(ctx.HeaderID, entry)
		allowed = false
	}()
	return entry.hook(ctx)
}

var _ plugin.PacketInterceptor = (*Interceptor)(nil)
