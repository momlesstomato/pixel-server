package sdk

// Priority controls handler execution order.
type Priority int

const (
	// PriorityLowest executes first in the handler chain.
	PriorityLowest Priority = 0
	// PriorityLow executes early in the handler chain.
	PriorityLow Priority = 25
	// PriorityNormal is the default execution priority.
	PriorityNormal Priority = 50
	// PriorityHigh executes late in the handler chain.
	PriorityHigh Priority = 75
	// PriorityHighest executes last among regular handlers.
	PriorityHighest Priority = 100
	// PriorityMonitor always executes regardless of cancellation.
	PriorityMonitor Priority = 127
)

// HandlerOption configures event handler behavior.
type HandlerOption func(*HandlerConfig)

// HandlerConfig stores resolved handler configuration.
type HandlerConfig struct {
	// Priority stores handler execution priority.
	Priority Priority
	// SkipCancelled stores whether to skip on cancelled events.
	SkipCancelled bool
}

// WithPriority sets the handler execution priority.
func WithPriority(p Priority) HandlerOption {
	return func(c *HandlerConfig) { c.Priority = p }
}

// SkipCancelled causes the handler to be skipped if the event is already cancelled.
func SkipCancelled() HandlerOption {
	return func(c *HandlerConfig) { c.SkipCancelled = true }
}

// EventBus allows subscribing to typed events.
type EventBus interface {
	// Subscribe registers a handler for events of type T.
	Subscribe(handler any, opts ...HandlerOption) func()
}
