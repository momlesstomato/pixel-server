package migrations

import "testing"

// TestRegistryReturnsOrderedSteps verifies migration step registry contents.
func TestRegistryReturnsOrderedSteps(t *testing.T) {
	steps := Registry()
	if len(steps) != 38 {
		t.Fatalf("expected 38 migration steps, got %d", len(steps))
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
}
