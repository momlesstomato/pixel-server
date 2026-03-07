package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"pixelsv/pkg/plugin"
)

// TestDiscoverLoadsSharedObjects validates extension filtering and ordering.
func TestDiscoverLoadsSharedObjects(t *testing.T) {
	dir := t.TempDir()
	paths := []string{"b.so", "a.so", "README.txt"}
	for _, name := range paths {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	loaded := make([]string, 0)
	opener := func(path string) (plugin.Plugin, error) {
		loaded = append(loaded, filepath.Base(path))
		return newFakePlugin(filepath.Base(path)), nil
	}
	plugins, err := Discover(dir, opener)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected two plugins, got %d", len(plugins))
	}
	if loaded[0] != "a.so" || loaded[1] != "b.so" {
		t.Fatalf("unexpected load order: %v", loaded)
	}
}

// TestDiscoverBubblesLoaderError validates opener error propagation.
func TestDiscoverBubblesLoaderError(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "broken.so")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	opener := func(path string) (plugin.Plugin, error) {
		return nil, fmt.Errorf("bad plugin")
	}
	_, err := Discover(dir, opener)
	if err == nil {
		t.Fatalf("expected error")
	}
}
