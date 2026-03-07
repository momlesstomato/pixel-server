package loader

import (
	"testing"

	"pixelsv/pkg/plugin"
)

// TestNewBuildsDependencyOrder validates sorted registry creation.
func TestNewBuildsDependencyOrder(t *testing.T) {
	a := newFakePlugin("a")
	b := newFakePlugin("b", "a")
	registry, err := New([]plugin.Plugin{b, a}, []string{"auth"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	plugins := registry.Plugins()
	if len(plugins) != 2 {
		t.Fatalf("expected two plugins, got %d", len(plugins))
	}
	if plugins[0].Metadata().Name != "a" || plugins[1].Metadata().Name != "b" {
		t.Fatalf("unexpected order: %s, %s", plugins[0].Metadata().Name, plugins[1].Metadata().Name)
	}
}

// TestEnableAllSkipsRealmMismatch validates realm-scoped activation.
func TestEnableAllSkipsRealmMismatch(t *testing.T) {
	a := newFakePlugin("auth-only")
	a.metadata.Realms = []string{"auth"}
	registry, err := New([]plugin.Plugin{a}, []string{"gateway"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = registry.EnableAll(func(metadata plugin.Metadata) plugin.API {
		return fakeAPI{}
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	status := registry.Status()
	if len(status) != 1 || status[0].State != PluginStateSkipped {
		t.Fatalf("unexpected status: %+v", status)
	}
	if a.enabled != 0 {
		t.Fatalf("expected plugin to be skipped")
	}
}
