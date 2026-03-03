// Package intercept provides the packet interception pipeline for the plugin
// system. Hooks run before and after default handlers in registration order.
package intercept

import (
	"sync"

	"pixel-server/pkg/plugin/event"
)

// PacketContext is passed to every hook function and describes the packet
// being processed. Hooks may mutate Payload or set Cancel to true.
type PacketContext struct {
	// SessionID is the unique identifier of the player session.
	SessionID string

	// HeaderID is the packet identifier.
	HeaderID uint16

	// Payload is the raw packet body (after the header bytes).
	// Hooks may replace this slice to modify the packet in transit.
	Payload []byte

	// Cancel indicates the packet should be dropped entirely.
	Cancel bool

	// Direction is "c2s" (client to server) or "s2c" (server to client).
	Direction string
}

// HookFunc is a callback invoked when a packet with a matching headerID is processed.
type HookFunc func(ctx *PacketContext)

// Interceptor lets plugins hook into the packet routing pipeline.
type Interceptor interface {
	// Before registers a hook that runs before the default handler.
	Before(headerID uint16, fn HookFunc) event.CancelFunc

	// After registers a hook that runs after the default handler.
	After(headerID uint16, fn HookFunc) event.CancelFunc

	// RunBefore executes all Before hooks for the context header in registration order.
	RunBefore(ctx *PacketContext)

	// RunAfter executes all After hooks for the context header in registration order.
	RunAfter(ctx *PacketContext)
}

type hookEntry struct {
	id uint64
	fn HookFunc
}

type hookChain struct {
	mu      sync.RWMutex
	entries []hookEntry
	counter uint64
}

func (c *hookChain) add(fn HookFunc) event.CancelFunc {
	c.mu.Lock()
	c.counter++
	id := c.counter
	c.entries = append(c.entries, hookEntry{id: id, fn: fn})
	c.mu.Unlock()

	return func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		for i, e := range c.entries {
			if e.id == id {
				c.entries = append(c.entries[:i], c.entries[i+1:]...)
				return
			}
		}
	}
}

func (c *hookChain) run(ctx *PacketContext) {
	c.mu.RLock()
	entries := make([]hookEntry, len(c.entries))
	copy(entries, c.entries)
	c.mu.RUnlock()

	for _, e := range entries {
		e.fn(ctx)
		if ctx.Cancel {
			return
		}
	}
}

type chainInterceptor struct {
	mu     sync.RWMutex
	before map[uint16]*hookChain
	after  map[uint16]*hookChain
}

// NewInterceptor creates a new packet interceptor.
func NewInterceptor() Interceptor {
	return &chainInterceptor{
		before: make(map[uint16]*hookChain),
		after:  make(map[uint16]*hookChain),
	}
}

func (ci *chainInterceptor) chain(m map[uint16]*hookChain, headerID uint16) *hookChain {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	c, ok := m[headerID]
	if !ok {
		c = &hookChain{}
		m[headerID] = c
	}
	return c
}

func (ci *chainInterceptor) Before(headerID uint16, fn HookFunc) event.CancelFunc {
	return ci.chain(ci.before, headerID).add(fn)
}

func (ci *chainInterceptor) After(headerID uint16, fn HookFunc) event.CancelFunc {
	return ci.chain(ci.after, headerID).add(fn)
}

func (ci *chainInterceptor) RunBefore(ctx *PacketContext) {
	ci.mu.RLock()
	c := ci.before[ctx.HeaderID]
	ci.mu.RUnlock()
	if c != nil {
		c.run(ctx)
	}
}

func (ci *chainInterceptor) RunAfter(ctx *PacketContext) {
	ci.mu.RLock()
	c := ci.after[ctx.HeaderID]
	ci.mu.RUnlock()
	if c != nil {
		c.run(ctx)
	}
}
