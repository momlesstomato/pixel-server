package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 3 {
		t.Fatalf("expected three migration steps, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260312_01_system_settings" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[1] == nil || steps[1].ID != "20260312_02_users" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[2] == nil || steps[2].ID != "20260312_03_system_settings_audit_owner" {
		t.Fatalf("unexpected migration step metadata")
	}
}
