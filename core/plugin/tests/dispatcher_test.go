package tests

import (
	"testing"

	sdk "github.com/momlesstomato/pixel-sdk"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	"go.uber.org/zap"
)

// TestDispatcherPriorityOrdering verifies handlers execute in priority order.
func TestDispatcherPriorityOrdering(t *testing.T) {
	d := coreplugin.NewDispatcher(zap.NewNop())
	var order []string
	d.Subscribe("test", func(e *sdk.AuthCompleted) { order = append(order, "high") }, sdk.WithPriority(sdk.PriorityHigh))
	d.Subscribe("test", func(e *sdk.AuthCompleted) { order = append(order, "low") }, sdk.WithPriority(sdk.PriorityLow))
	d.Subscribe("test", func(e *sdk.AuthCompleted) { order = append(order, "normal") })
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if len(order) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(order))
	}
	if order[0] != "low" || order[1] != "normal" || order[2] != "high" {
		t.Errorf("unexpected order: %v", order)
	}
}

// TestDispatcherCancellationChain verifies cancellation propagation.
func TestDispatcherCancellationChain(t *testing.T) {
	d := coreplugin.NewDispatcher(zap.NewNop())
	var executed []string
	d.Subscribe("test", func(e *sdk.AuthValidating) {
		e.Cancel()
		executed = append(executed, "canceller")
	}, sdk.WithPriority(sdk.PriorityLow))
	d.Subscribe("test", func(e *sdk.AuthValidating) {
		executed = append(executed, "skipped")
	}, sdk.WithPriority(sdk.PriorityNormal), sdk.SkipCancelled())
	d.Subscribe("test", func(e *sdk.AuthValidating) {
		executed = append(executed, "monitor")
	}, sdk.WithPriority(sdk.PriorityMonitor))
	event := &sdk.AuthValidating{ConnID: "c1", UserID: 1, Ticket: "t"}
	d.Fire(event)
	if !event.Cancelled() {
		t.Error("event should be cancelled")
	}
	if len(executed) != 2 {
		t.Fatalf("expected 2 executed, got %d: %v", len(executed), executed)
	}
	if executed[0] != "canceller" || executed[1] != "monitor" {
		t.Errorf("unexpected execution: %v", executed)
	}
}

// TestDispatcherPanicRecovery verifies panicking handler does not crash chain.
func TestDispatcherPanicRecovery(t *testing.T) {
	d := coreplugin.NewDispatcher(zap.NewNop())
	var reached bool
	d.Subscribe("test", func(e *sdk.AuthCompleted) { panic("boom") }, sdk.WithPriority(sdk.PriorityLow))
	d.Subscribe("test", func(e *sdk.AuthCompleted) { reached = true }, sdk.WithPriority(sdk.PriorityNormal))
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if !reached {
		t.Error("second handler should have executed after panic")
	}
}

// TestDispatcherUnsubscribe verifies handler removal.
func TestDispatcherUnsubscribe(t *testing.T) {
	d := coreplugin.NewDispatcher(zap.NewNop())
	var count int
	unsub := d.Subscribe("test", func(e *sdk.AuthCompleted) { count++ })
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if count != 1 {
		t.Fatalf("expected 1 call, got %d", count)
	}
	unsub()
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if count != 1 {
		t.Errorf("expected 1 call after unsubscribe, got %d", count)
	}
}

// TestDispatcherRemoveByOwner verifies owner-based cleanup.
func TestDispatcherRemoveByOwner(t *testing.T) {
	d := coreplugin.NewDispatcher(zap.NewNop())
	var pluginCount, coreCount int
	d.Subscribe("my-plugin", func(e *sdk.AuthCompleted) { pluginCount++ })
	d.Subscribe("core", func(e *sdk.AuthCompleted) { coreCount++ })
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if pluginCount != 1 || coreCount != 1 {
		t.Fatalf("expected both to fire: plugin=%d core=%d", pluginCount, coreCount)
	}
	d.RemoveByOwner("my-plugin")
	d.Fire(&sdk.AuthCompleted{ConnID: "c1", UserID: 1})
	if pluginCount != 1 {
		t.Errorf("plugin handler should not fire after removal: %d", pluginCount)
	}
	if coreCount != 2 {
		t.Errorf("core handler should still fire: %d", coreCount)
	}
}
