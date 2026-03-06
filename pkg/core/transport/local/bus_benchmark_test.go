package local

import (
	"context"
	"testing"

	"pixelsv/pkg/core/transport"
)

// BenchmarkBusPublish measures local bus publish overhead for 20Hz-sensitive paths.
func BenchmarkBusPublish(b *testing.B) {
	bus := New()
	ctx := context.Background()
	_, err := bus.Subscribe(ctx, "room.input.*", func(context.Context, transport.Message) error {
		return nil
	})
	if err != nil {
		b.Fatalf("expected no error, got %v", err)
	}
	payload := []byte("payload")
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		if err := bus.Publish(ctx, "room.input.1", payload); err != nil {
			b.Fatalf("expected no error, got %v", err)
		}
	}
}
