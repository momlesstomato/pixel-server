package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 9 {
		t.Fatalf("expected nine migration steps, got %d", len(steps))
	}
	if steps[0] == nil || steps[0].ID != "20260314_01_permission_groups" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[1] == nil || steps[1].ID != "20260314_02_group_permissions" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[2] == nil || steps[2].ID != "20260312_01_users" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[3] == nil || steps[3].ID != "20260314_03_user_permission_groups" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[4] == nil || steps[4].ID != "20260312_02_user_login_events" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[5] == nil || steps[5].ID != "20260313_03_user_settings" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[6] == nil || steps[6].ID != "20260313_04_user_respects" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[7] == nil || steps[7].ID != "20260313_05_user_wardrobe" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[8] == nil || steps[8].ID != "20260313_06_user_ignores" {
		t.Fatalf("unexpected migration step metadata")
	}
}
