package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 47 {
		t.Fatalf("expected 47 migration steps, got %d", len(steps))
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
	if steps[3] == nil || steps[3].ID != "20260314_04_users_table_rename" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[4] == nil || steps[4].ID != "20260314_03_user_permission_groups" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[5] == nil || steps[5].ID != "20260312_02_user_login_events" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[6] == nil || steps[6].ID != "20260313_03_user_settings" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[7] == nil || steps[7].ID != "20260313_04_user_respects" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[8] == nil || steps[8].ID != "20260313_05_user_wardrobe" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[9] == nil || steps[9].ID != "20260313_06_user_ignores" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[10] == nil || steps[10].ID != "20260315_01_messenger_friendships" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[11] == nil || steps[11].ID != "20260315_02_friend_requests" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[12] == nil || steps[12].ID != "20260315_03_offline_messages" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[13] == nil || steps[13].ID != "20260315_04_normalize_messenger_friendships" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[14] == nil || steps[14].ID != "20260315_05_messenger_message_log" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[17] == nil || steps[17].ID != "20260401_01_drop_revision" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[18] == nil || steps[18].ID != "20260401_02_restore_sprite_id" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[31] == nil || steps[31].ID != "20260324_09_drop_cost_primary_type" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[32] == nil || steps[32].ID != "20260324_10_rename_cost_columns" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[36] == nil || steps[36].ID != "20260320_13_subscriptions" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[37] == nil || steps[37].ID != "20260325_14_club_offers" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[38] == nil || steps[38].ID != "20260325_13_navigator_categories" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[39] == nil || steps[39].ID != "20260325_14_rooms" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[40] == nil || steps[40].ID != "20260325_15_navigator_saved_searches" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[41] == nil || steps[41].ID != "20260325_16_navigator_favourites" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[42] == nil || steps[42].ID != "20260401_01_room_models" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[43] == nil || steps[43].ID != "20260401_02_room_extension" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[44] == nil || steps[44].ID != "20260401_03_room_bans" {
		t.Fatalf("unexpected migration step metadata")
	}
	if steps[45] == nil || steps[45].ID != "20260401_04_room_rights" {
		t.Fatalf("unexpected migration step metadata")
	}
}
