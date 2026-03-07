package loader

import (
	"strings"
	"testing"

	"pixelsv/pkg/plugin"
)

// TestSortByDependenciesOrdersByDependency validates topological ordering.
func TestSortByDependenciesOrdersByDependency(t *testing.T) {
	a := newFakePlugin("a")
	b := newFakePlugin("b", "a")
	c := newFakePlugin("c", "b")
	ordered, err := SortByDependencies([]plugin.Plugin{c, b, a})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ordered) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(ordered))
	}
	if ordered[0].Metadata().Name != "a" || ordered[1].Metadata().Name != "b" || ordered[2].Metadata().Name != "c" {
		t.Fatalf("unexpected order: %s, %s, %s", ordered[0].Metadata().Name, ordered[1].Metadata().Name, ordered[2].Metadata().Name)
	}
}

// TestSortByDependenciesMissingDependency validates missing dependency failures.
func TestSortByDependenciesMissingDependency(t *testing.T) {
	a := newFakePlugin("a", "missing")
	_, err := SortByDependencies([]plugin.Plugin{a})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestSortByDependenciesCycle validates cycle detection.
func TestSortByDependenciesCycle(t *testing.T) {
	a := newFakePlugin("a", "b")
	b := newFakePlugin("b", "a")
	_, err := SortByDependencies([]plugin.Plugin{a, b})
	if err == nil {
		t.Fatalf("expected cycle error")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("unexpected error: %v", err)
	}
}
