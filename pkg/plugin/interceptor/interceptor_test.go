package interceptor

import (
	"testing"

	"pixelsv/pkg/plugin"
)

// TestInterceptorRunBeforeCancellation validates cancellation behavior.
func TestInterceptorRunBeforeCancellation(t *testing.T) {
	i := New()
	i.BeforeAll(func(ctx plugin.PacketContext) bool {
		return false
	})
	allowed := i.RunBefore(plugin.PacketContext{HeaderID: 4000})
	if allowed {
		t.Fatalf("expected cancelled packet")
	}
}

// TestInterceptorRunAfterCallsHook validates post-hook dispatch.
func TestInterceptorRunAfterCallsHook(t *testing.T) {
	i := New()
	calls := 0
	i.After(100, func(ctx plugin.PacketContext) bool {
		calls++
		return true
	})
	i.RunAfter(plugin.PacketContext{HeaderID: 100})
	if calls != 1 {
		t.Fatalf("expected one call, got %d", calls)
	}
}

// TestInterceptorUnsubscribeStopsHook validates unregistration.
func TestInterceptorUnsubscribeStopsHook(t *testing.T) {
	i := New()
	calls := 0
	reg := i.Before(200, func(ctx plugin.PacketContext) bool {
		calls++
		return true
	})
	reg.Unsubscribe()
	allowed := i.RunBefore(plugin.PacketContext{HeaderID: 200})
	if !allowed {
		t.Fatalf("expected allowed packet")
	}
	if calls != 0 {
		t.Fatalf("expected zero calls, got %d", calls)
	}
}

// TestInterceptorPanicRecoveryRemovesHook validates panic isolation and cleanup.
func TestInterceptorPanicRecoveryRemovesHook(t *testing.T) {
	panicCalls := 0
	i := NewWithPanicHandler(func(headerID uint16, recovered any) {
		if headerID == 300 && recovered != nil {
			panicCalls++
		}
	})
	calls := 0
	i.Before(300, func(ctx plugin.PacketContext) bool {
		calls++
		panic("panic hook")
	})
	allowed := i.RunBefore(plugin.PacketContext{HeaderID: 300})
	if allowed {
		t.Fatalf("expected panic to cancel packet")
	}
	if panicCalls != 1 {
		t.Fatalf("expected one panic callback, got %d", panicCalls)
	}
	if calls != 1 {
		t.Fatalf("expected one call before removal, got %d", calls)
	}
	allowed = i.RunBefore(plugin.PacketContext{HeaderID: 300})
	if !allowed {
		t.Fatalf("expected no hooks after panic cleanup")
	}
}
