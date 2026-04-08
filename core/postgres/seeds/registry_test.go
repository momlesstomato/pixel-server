package seeds

import "testing"

// TestRegistryReturnsOrderedSeedSteps verifies seed step registry contents.
func TestRegistryReturnsOrderedSeedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 28 {
		t.Fatalf("expected twenty-eight seed steps, got %d", len(steps))
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
	if steps[8] == nil || steps[8].ID != "20260405_03_staff_ambassador_groups" {
		t.Fatalf("unexpected seed step metadata at index 8")
	}
	if steps[9] == nil || steps[9].ID != "20260405_04_staff_ambassador_permissions" {
		t.Fatalf("unexpected seed step metadata at index 9")
	}
	if steps[10] == nil || steps[10].ID != "20260405_07_extended_group_users" {
		t.Fatalf("unexpected seed step metadata at index 10")
	}
	if steps[11] == nil || steps[11].ID != "20260405_08_extended_group_user_settings" {
		t.Fatalf("unexpected seed step metadata at index 11")
	}
	if steps[12] == nil || steps[12].ID != "20260406_05_security_level_backfill" {
		t.Fatalf("unexpected seed step metadata at index 12")
	}
	if steps[13] == nil || steps[13].ID != "20260324_inv_01_currency_types" {
		t.Fatalf("unexpected seed step metadata at index 13")
	}
	if steps[14] == nil || steps[14].ID != "20260320_S01_catalog_pages" {
		t.Fatalf("unexpected seed step metadata at index 14")
	}
	if steps[15] == nil || steps[15].ID != "20260405_S02_hc_shop_page" {
		t.Fatalf("unexpected seed step metadata at index 15")
	}
	if steps[16] == nil || steps[16].ID != "20260406_S04_hc_shop_localization_backfill" {
		t.Fatalf("unexpected seed step metadata at index 16")
	}
	if steps[17] == nil || steps[17].ID != "20260406_S05_club_gifts_page" {
		t.Fatalf("unexpected seed step metadata at index 17")
	}
	if steps[18] == nil || steps[18].ID != "20260407_S08_hc_shop_vip_buy_backfill" {
		t.Fatalf("unexpected seed step metadata at index 18")
	}
	if steps[19] == nil || steps[19].ID != "20260320_S02_club_offers" {
		t.Fatalf("unexpected seed step metadata at index 19")
	}
	if steps[20] == nil || steps[20].ID != "20260405_S03_subscription_users" {
		t.Fatalf("unexpected seed step metadata at index 20")
	}
	if steps[21] == nil || steps[21].ID != "20260406_S06_default_club_gifts" {
		t.Fatalf("unexpected seed step metadata at index 21")
	}
	if steps[22] == nil || steps[22].ID != "20260406_S07_default_payday_config" {
		t.Fatalf("unexpected seed step metadata at index 22")
	}
	if steps[23] == nil || steps[23].ID != "20260326_S01_nav_categories" {
		t.Fatalf("unexpected seed step metadata at index 23")
	}
	if steps[24] == nil || steps[24].ID != "20260326_S02_nav_demo_rooms" {
		t.Fatalf("unexpected seed step metadata at index 24")
	}
	if steps[25] == nil || steps[25].ID != "20260408_S03_nav_demo_admin_room_owner_backfill" {
		t.Fatalf("unexpected seed step metadata at index 25")
	}
	if steps[26] == nil || steps[26].ID != "seed_20260401_01_room_models" {
		t.Fatalf("unexpected seed step metadata at index 26")
	}
	if steps[27] == nil || steps[27].ID != "20260405_09_permission_assignment_backfill" {
		t.Fatalf("unexpected seed step metadata at index 27")
	}
}
