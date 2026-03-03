package plugin

import (
	"fmt"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"pixel-server/pkg/plugin/event"
)

type entry struct {
	plugin  Plugin
	meta    Meta
	enabled bool
	cancels []event.CancelFunc
}

// APIProvider creates a plugin-scoped API instance.
type APIProvider interface {
	// PluginAPI returns an API scoped to pluginName.
	PluginAPI(pluginName string, configDir string) API
}

// Registry manages plugin discovery and lifecycle.
type Registry struct {
	mu          sync.Mutex
	plugins     []*entry
	apiProvider APIProvider
	pluginsDir  string
	log         *zap.Logger
}

// NewRegistry constructs a registry.
func NewRegistry(provider APIProvider, pluginsDir string, log *zap.Logger) *Registry {
	return &Registry{apiProvider: provider, pluginsDir: pluginsDir, log: log}
}

// LoadAll loads every plugin binary from the configured directory and enables them.
func (r *Registry) LoadAll() error {
	paths := collectPluginPaths(r.pluginsDir)

	r.mu.Lock()
	for _, path := range paths {
		loaded, err := LoadFromFile(path)
		if err != nil {
			r.log.Warn("failed to load plugin", zap.String("path", path), zap.Error(err))
			continue
		}
		meta := loaded.Meta()
		r.plugins = append(r.plugins, &entry{plugin: loaded, meta: meta})
		r.log.Info("loaded plugin", zap.String("name", meta.Name), zap.String("version", meta.Version), zap.String("path", path))
	}
	r.sortByDependencies()
	r.mu.Unlock()

	return r.EnableAll()
}

func collectPluginPaths(pluginsDir string) []string {
	soFiles, _ := filepath.Glob(filepath.Join(pluginsDir, "*.so"))
	dyFiles, _ := filepath.Glob(filepath.Join(pluginsDir, "*.dylib"))
	return append(soFiles, dyFiles...)
}

// EnableAll enables all loaded plugins in dependency order.
func (r *Registry) EnableAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, pluginEntry := range r.plugins {
		if pluginEntry.enabled {
			continue
		}
		if err := r.enableLocked(pluginEntry); err != nil {
			r.log.Error("failed to enable plugin", zap.String("name", pluginEntry.meta.Name), zap.Error(err))
		}
	}
	return nil
}

func (r *Registry) enableLocked(pluginEntry *entry) (retErr error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = fmt.Errorf("panic in %s.OnEnable: %v", pluginEntry.meta.Name, recovered)
		}
	}()

	base := r.apiProvider.PluginAPI(pluginEntry.meta.Name, r.pluginsDir)
	api := wrapAPI(base, func(cancel event.CancelFunc) {
		pluginEntry.cancels = append(pluginEntry.cancels, cancel)
	})
	if err := pluginEntry.plugin.OnEnable(api); err != nil {
		return fmt.Errorf("%s.OnEnable: %w", pluginEntry.meta.Name, err)
	}

	pluginEntry.enabled = true
	r.log.Info("enabled plugin", zap.String("name", pluginEntry.meta.Name))
	return nil
}

// DisableAll disables all enabled plugins in reverse order.
func (r *Registry) DisableAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := len(r.plugins) - 1; i >= 0; i-- {
		pluginEntry := r.plugins[i]
		if !pluginEntry.enabled {
			continue
		}
		r.disableLocked(pluginEntry)
	}
}

func (r *Registry) disableLocked(pluginEntry *entry) {
	defer func() {
		if recovered := recover(); recovered != nil {
			r.log.Error("panic in plugin OnDisable", zap.String("name", pluginEntry.meta.Name), zap.Any("panic", recovered))
		}
	}()

	for _, cancel := range pluginEntry.cancels {
		cancel()
	}
	pluginEntry.cancels = nil

	if err := pluginEntry.plugin.OnDisable(); err != nil {
		r.log.Error("plugin OnDisable error", zap.String("name", pluginEntry.meta.Name), zap.Error(err))
	}
	pluginEntry.enabled = false
	r.log.Info("disabled plugin", zap.String("name", pluginEntry.meta.Name))
}

// InjectPlugin inserts a pre-built plugin instance. Intended for tests only.
func (r *Registry) InjectPlugin(p Plugin) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.plugins = append(r.plugins, &entry{plugin: p, meta: p.Meta()})
}

// List returns metadata for all loaded plugins.
func (r *Registry) List() []Meta {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Meta, len(r.plugins))
	for i, pluginEntry := range r.plugins {
		out[i] = pluginEntry.meta
	}
	return out
}
