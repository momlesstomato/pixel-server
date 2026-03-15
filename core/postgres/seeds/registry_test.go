package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 5 {
		t.Fatalf("expected five seed steps, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260313_01_system_user" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[1] == nil || steps[1].ID != "20260313_02_system_settings" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[2] == nil || steps[2].ID != "20260314_01_default_permission_groups" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[3] == nil || steps[3].ID != "20260314_02_default_group_permissions" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[4] == nil || steps[4].ID != "20260315_03_test_users" {
		t.Fatalf("unexpected seed step metadata")
	}
}
