package plugin

import "testing"

// TestEventCancelMarksEvent validates cancellation state transition.
func TestEventCancelMarksEvent(t *testing.T) {
	event := &Event{Name: "room.user.join"}
	if event.Cancelled() {
		t.Fatalf("expected default not cancelled")
	}
	event.Cancel()
	if !event.Cancelled() {
		t.Fatalf("expected cancelled event")
	}
}

// TestEventNilSafety validates nil receiver behavior.
func TestEventNilSafety(t *testing.T) {
	var event *Event
	event.Cancel()
	if event.Cancelled() {
		t.Fatalf("expected nil event to report not cancelled")
	}
}
