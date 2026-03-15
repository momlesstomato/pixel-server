package permission

import "testing"

// TestResolveExactMatch verifies direct permission lookup behavior.
func TestResolveExactMatch(t *testing.T) {
	grants := map[string]struct{}{"moderation.kick": {}}
	if !Resolve(grants, "moderation.kick") {
		t.Fatalf("expected exact match to resolve")
	}
}

// TestResolveHierarchyWildcard verifies dotted wildcard resolution behavior.
func TestResolveHierarchyWildcard(t *testing.T) {
	grants := map[string]struct{}{"moderation.*": {}}
	if !Resolve(grants, "moderation.ban") {
		t.Fatalf("expected wildcard match to resolve")
	}
}

// TestResolveRootWildcard verifies root wildcard resolution behavior.
func TestResolveRootWildcard(t *testing.T) {
	grants := map[string]struct{}{WildcardPermission: {}}
	if !Resolve(grants, "any.permission") {
		t.Fatalf("expected root wildcard to resolve")
	}
}

// TestResolveNestedHierarchy verifies deep dotted resolution behavior.
func TestResolveNestedHierarchy(t *testing.T) {
	grants := map[string]struct{}{"a.b.*": {}}
	if !Resolve(grants, "a.b.c.d") {
		t.Fatalf("expected nested wildcard to resolve")
	}
}

// TestResolveRejectsUnknown verifies unknown permission behavior.
func TestResolveRejectsUnknown(t *testing.T) {
	grants := map[string]struct{}{"a.b": {}}
	if Resolve(grants, "a.c") {
		t.Fatalf("expected unknown permission to fail")
	}
}

// TestResolveRejectsEmpty verifies empty permission behavior.
func TestResolveRejectsEmpty(t *testing.T) {
	if Resolve(map[string]struct{}{"a.b": {}}, " ") {
		t.Fatalf("expected empty permission to fail")
	}
}
