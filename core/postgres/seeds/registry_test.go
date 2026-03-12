package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 1 {
		t.Fatalf("expected one seed step, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260312_01_seed_system_settings" {
		t.Fatalf("unexpected seed step metadata")
	}
}
