package loader

import (
	"fmt"
	"os"
	"path/filepath"
	stdplugin "plugin"
	"sort"
	"strings"

	"pixelsv/pkg/plugin"
)

// Discover loads all plugin shared objects found in one directory.
func Discover(directory string, opener BinaryOpener) ([]plugin.Plugin, error) {
	if opener == nil {
		opener = OpenSharedObject
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".so") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	result := make([]plugin.Plugin, 0, len(paths))
	for _, path := range paths {
		instance, loadErr := opener(path)
		if loadErr != nil {
			return nil, fmt.Errorf("load plugin %s: %w", path, loadErr)
		}
		result = append(result, instance)
	}
	return result, nil
}

// OpenSharedObject loads one plugin instance from a .so file.
func OpenSharedObject(path string) (plugin.Plugin, error) {
	handle, err := stdplugin.Open(path)
	if err != nil {
		return nil, err
	}
	symbol, err := handle.Lookup("NewPlugin")
	if err != nil {
		return nil, err
	}
	factory, err := normalizeFactory(symbol)
	if err != nil {
		return nil, err
	}
	instance := factory()
	if instance == nil {
		return nil, fmt.Errorf("NewPlugin returned nil instance")
	}
	return instance, nil
}

// normalizeFactory converts lookup symbols into one supported factory signature.
func normalizeFactory(symbol any) (func() plugin.Plugin, error) {
	if direct, ok := symbol.(func() plugin.Plugin); ok {
		return direct, nil
	}
	if ref, ok := symbol.(*func() plugin.Plugin); ok && ref != nil {
		return *ref, nil
	}
	return nil, fmt.Errorf("NewPlugin symbol has unsupported type %T", symbol)
}
