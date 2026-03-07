package loader

import (
	"errors"
	"strings"
	"testing"

	"pixelsv/pkg/plugin"
)

// TestEnableAllSkipsWhenDependencyFailed validates dependency gating.
func TestEnableAllSkipsWhenDependencyFailed(t *testing.T) {
	dep := newFakePlugin("dep")
	dep.enableErr = errors.New("enable fail")
	main := newFakePlugin("main", "dep")
	registry, err := New([]plugin.Plugin{main, dep}, []string{"auth"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err = registry.EnableAll(func(metadata plugin.Metadata) plugin.API {
		return fakeAPI{}
	})
	if err == nil {
		t.Fatalf("expected combined error")
	}
	status := registry.Status()
	if status[0].State != PluginStateFailed {
		t.Fatalf("expected dep failed, got %+v", status[0])
	}
	if status[1].State != PluginStateSkipped {
		t.Fatalf("expected main skipped, got %+v", status[1])
	}
	if !strings.Contains(status[1].Err.Error(), "dependency dep is not enabled") {
		t.Fatalf("unexpected dependency error: %v", status[1].Err)
	}
}

// TestDisableAllRunsReverseOrder validates reverse disable lifecycle.
func TestDisableAllRunsReverseOrder(t *testing.T) {
	a := newFakePlugin("a")
	b := newFakePlugin("b", "a")
	registry, err := New([]plugin.Plugin{b, a}, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := registry.EnableAll(func(metadata plugin.Metadata) plugin.API { return fakeAPI{} }); err != nil {
		t.Fatalf("expected no enable error, got %v", err)
	}
	if err := registry.DisableAll(); err != nil {
		t.Fatalf("expected no disable error, got %v", err)
	}
	if a.disabled != 1 || b.disabled != 1 {
		t.Fatalf("expected both plugins disabled once: a=%d b=%d", a.disabled, b.disabled)
	}
	status := registry.Status()
	if status[0].State != PluginStateDisabled || status[1].State != PluginStateDisabled {
		t.Fatalf("unexpected status: %+v", status)
	}
}
