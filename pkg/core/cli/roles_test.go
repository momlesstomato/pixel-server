package cli

import "testing"

// TestNewRoleSet validates role parsing into a set.
func TestNewRoleSet(t *testing.T) {
	roles, err := newRoleSet("gateway,api,gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !roles.has("gateway") || !roles.has("api") {
		t.Fatalf("expected gateway and api roles")
	}
	if len(roles) != 2 {
		t.Fatalf("expected deduplicated role set, got %d", len(roles))
	}
}

// TestRoleSetNeeds validates dependency requirements by role.
func TestRoleSetNeeds(t *testing.T) {
	gateway, err := newRoleSet("gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !gateway.needsHTTP() || gateway.needsPostgres() || !gateway.needsRedis() {
		t.Fatalf("unexpected gateway dependency plan")
	}
	game, err := newRoleSet("game")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if game.needsHTTP() || !game.needsPostgres() || !game.needsRedis() {
		t.Fatalf("unexpected game dependency plan")
	}
	all, err := newRoleSet("all")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !all.needsHTTP() || !all.needsPostgres() || !all.needsRedis() {
		t.Fatalf("unexpected all dependency plan")
	}
	if !all.forceLocalTransport() {
		t.Fatalf("expected all role to force local transport")
	}
	if gateway.forceLocalTransport() {
		t.Fatalf("expected gateway role to allow distributed transport")
	}
}

// TestRoleSetNames validates sorted role name rendering.
func TestRoleSetNames(t *testing.T) {
	roles, err := newRoleSet("moderation,api,gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	names := roles.names()
	if len(names) != 3 {
		t.Fatalf("unexpected role names length: %d", len(names))
	}
	if names[0] != "api" || names[1] != "gateway" || names[2] != "moderation" {
		t.Fatalf("unexpected role names: %+v", names)
	}
}
