//go:build linux || darwin

package plugin

import (
	"fmt"
	"plugin"
)

// LoadFromFile opens a compiled plugin .so file at path and returns the
// Plugin instance by calling the mandatory exported NewPlugin() symbol.
//
// The .so must have been compiled with:
//
//	go build -buildmode=plugin -o myplugin.so ./myplugin/
//
// and must export exactly:
//
//	func NewPlugin() plugin.Plugin
//
// This function is only available on Linux and macOS (go:build linux || darwin).
func LoadFromFile(path string) (Plugin, error) {
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("plugin.Open(%q): %w", path, err)
	}

	sym, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %q: missing exported symbol NewPlugin: %w", path, err)
	}

	newFn, ok := sym.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("plugin %q: NewPlugin has wrong signature (want func() plugin.Plugin)", path)
	}

	return newFn(), nil
}
