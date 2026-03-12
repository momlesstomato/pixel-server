package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 1 {
		t.Fatalf("expected one migration step, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260312_01_system_settings" {
		t.Fatalf("unexpected migration step metadata")
	}
}
