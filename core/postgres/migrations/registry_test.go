package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 2 {
		t.Fatalf("expected two migration steps, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260312_01_users" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[1] == nil || steps[1].ID != "20260312_02_user_login_events" {
		t.Fatalf("unexpected migration step metadata")
	}
}
