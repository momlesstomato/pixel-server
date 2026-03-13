package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 2 {
		t.Fatalf("expected two seed steps, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260313_01_system_user" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[1] == nil || steps[1].ID != "20260313_02_system_settings" {
		t.Fatalf("unexpected seed step metadata")
	}
}
