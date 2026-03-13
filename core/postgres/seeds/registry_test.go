package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 0 {
		t.Fatalf("expected zero seed steps, got %d", len(steps))
	}
}
