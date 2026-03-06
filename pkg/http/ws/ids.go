package ws

import (
	"strconv"
	"sync/atomic"
)

// SessionIDGenerator generates stable monotonic session ids.
type SessionIDGenerator struct {
	// next stores the next numeric sequence value.
	next atomic.Uint64
}

// NewSessionIDGenerator creates a new SessionIDGenerator.
func NewSessionIDGenerator() *SessionIDGenerator {
	return &SessionIDGenerator{}
}

// Next returns the next generated session id in base36.
func (g *SessionIDGenerator) Next() string {
	value := g.next.Add(1)
	return strconv.FormatUint(value, 36)
}
