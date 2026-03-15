package domain

import "testing"

// TestValidateGroupName verifies group-name validation behavior.
func TestValidateGroupName(t *testing.T) {
	valid, err := ValidateGroupName("Admin-Team")
	if err != nil {
		t.Fatalf("expected group name validation success, got %v", err)
	}
	if valid != "admin-team" {
		t.Fatalf("expected lowercase normalized name, got %q", valid)
	}
	if _, err := ValidateGroupName("A"); err == nil {
		t.Fatalf("expected short group name validation failure")
	}
}

// TestValidatePermission verifies permission string validation behavior.
func TestValidatePermission(t *testing.T) {
	valid, err := ValidatePermission("Moderation.Ban")
	if err != nil {
		t.Fatalf("expected permission validation success, got %v", err)
	}
	if valid != "moderation.ban" {
		t.Fatalf("expected lowercase normalized permission, got %q", valid)
	}
	if _, err := ValidatePermission("*.ban"); err == nil {
		t.Fatalf("expected invalid wildcard position failure")
	}
}
