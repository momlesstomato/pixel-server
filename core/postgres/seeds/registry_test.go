package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 14 {
		t.Fatalf("expected fourteen seed steps, got %d", len(steps))
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
	if steps[4] == nil || steps[4].ID != "20260314_03_test_users" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[5] == nil || steps[5].ID != "20260314_04_test_user_settings" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[6] == nil || steps[6].ID != "20260314_05_demo_users_backfill" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[7] == nil || steps[7].ID != "20260314_06_demo_user_settings_backfill" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[8] == nil || steps[8].ID != "20260324_inv_01_currency_types" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[9] == nil || steps[9].ID != "20260320_S01_catalog_pages" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[10] == nil || steps[10].ID != "20260320_S02_club_offers" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[11] == nil || steps[11].ID != "20260326_S01_nav_categories" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[12] == nil || steps[12].ID != "20260326_S02_nav_demo_rooms" {
		t.Fatalf("unexpected seed step metadata")
	}
	if steps[13] == nil || steps[13].ID != "seed_20260401_01_room_models" {
		t.Fatalf("unexpected seed step metadata")
	}
}
