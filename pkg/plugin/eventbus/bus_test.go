package eventbus

import (
	"errors"
	"testing"

	"pixelsv/pkg/plugin"
)

// TestBusEmitDispatchesHandlers validates single-event dispatch behavior.
func TestBusEmitDispatchesHandlers(t *testing.T) {
	bus := New()
	count := 0
	bus.On("auth.ticket.validated", func(event *plugin.Event) error {
		count++
		if event.Name != "auth.ticket.validated" {
			t.Fatalf("unexpected event name: %s", event.Name)
		}
		return nil
	})
	err := bus.Emit(&plugin.Event{Name: "auth.ticket.validated"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 call, got %d", count)
	}
}

// TestBusUnsubscribeStopsDispatch validates registration cleanup.
func TestBusUnsubscribeStopsDispatch(t *testing.T) {
	bus := New()
	count := 0
	reg := bus.On("session.connected", func(event *plugin.Event) error {
		count++
		return nil
	})
	reg.Unsubscribe()
	err := bus.Emit(&plugin.Event{Name: "session.connected"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no calls, got %d", count)
	}
}

// TestBusEmitValidation validates nil and missing-name errors.
func TestBusEmitValidation(t *testing.T) {
	bus := New()
	if err := bus.Emit(nil); !errors.Is(err, ErrNilEvent) {
		t.Fatalf("expected ErrNilEvent, got %v", err)
	}
	if err := bus.Emit(&plugin.Event{}); !errors.Is(err, ErrMissingEventName) {
		t.Fatalf("expected ErrMissingEventName, got %v", err)
	}
}

// TestBusPanicRecoveryRemovesHandler validates isolation and deregistration.
func TestBusPanicRecoveryRemovesHandler(t *testing.T) {
	panicCount := 0
	bus := NewWithPanicHandler(func(event string, recovered any) {
		if event == "room.tick.pre" && recovered != nil {
			panicCount++
		}
	})
	calls := 0
	bus.On("room.tick.pre", func(event *plugin.Event) error {
		calls++
		panic("boom")
	})
	if err := bus.Emit(&plugin.Event{Name: "room.tick.pre"}); err == nil {
		t.Fatalf("expected panic-derived error")
	}
	if panicCount != 1 {
		t.Fatalf("expected one panic callback, got %d", panicCount)
	}
	if calls != 1 {
		t.Fatalf("expected one handler call, got %d", calls)
	}
	if err := bus.Emit(&plugin.Event{Name: "room.tick.pre"}); err != nil {
		t.Fatalf("expected no handlers after panic cleanup, got %v", err)
	}
}
