package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sort"

	sdk "github.com/momlesstomato/pixel-sdk"
)

// PluginFactory defines the expected symbol type inside a .so file.
type PluginFactory = func() sdk.Plugin

// discoverPlugins scans a directory for .so files in alphabetical order.
func discoverPlugins(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".so" {
			paths = append(paths, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// loadPluginFactory opens a .so file and extracts the NewPlugin factory.
func loadPluginFactory(path string) (PluginFactory, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin open failed: %w", err)
	}
	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("symbol NewPlugin not found: %w", err)
	}
	factory, ok := sym.(*PluginFactory)
	if !ok {
		fn, fnOk := sym.(*func() sdk.Plugin)
		if !fnOk {
			return nil, fmt.Errorf("NewPlugin has wrong type: %T", sym)
		}
		return *fn, nil
	}
	return *factory, nil
}
