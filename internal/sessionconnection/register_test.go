package sessionconnection

import (
	"context"
	"testing"

	"pixelsv/pkg/core/transport/local"
	"pixelsv/pkg/plugin/eventbus"
)

// TestRegister validates realm registration lifecycle.
func TestRegister(t *testing.T) {
	bus := local.New()
	defer bus.Close()
	runtime, err := Register(context.Background(), bus, eventbus.New(), nil, DefaultConfig())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if runtime == nil || runtime.Service == nil {
		t.Fatalf("expected runtime service")
	}
}
