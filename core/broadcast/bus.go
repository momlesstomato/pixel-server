package broadcast

import (
	"context"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
)

// Broadcaster defines cross-instance publish and subscribe behavior.
type Broadcaster interface {
	// Publish sends one payload to one channel.
	Publish(context.Context, string, []byte) error
	// Subscribe listens for payloads in one channel.
	Subscribe(context.Context, string) (<-chan []byte, coreconnection.Disposable, error)
}
