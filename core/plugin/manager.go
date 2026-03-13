package plugin

import (
	"fmt"
	"path/filepath"

	sdk "github.com/momlesstomato/pixel-sdk"
	"go.uber.org/zap"
)

// loadedPlugin stores one enabled plugin with its metadata.
type loadedPlugin struct {
	plugin   sdk.Plugin
	manifest sdk.Manifest
	server   *serverImpl
}

// Manager manages plugin lifecycle: loading, enabling, and disabling.
type Manager struct {
	plugins    []loadedPlugin
	dispatcher *Dispatcher
	logger     *zap.Logger
	dir        string
}

// NewManager creates a plugin manager.
func NewManager(dispatcher *Dispatcher, logger *zap.Logger, dir string) (*Manager, error) {
	if dispatcher == nil {
		return nil, fmt.Errorf("dispatcher is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	if dir == "" {
		dir = "plugins"
	}
	return &Manager{dispatcher: dispatcher, logger: logger, dir: dir}, nil
}

// LoadAll scans the plugin directory and enables all discovered plugins.
func (m *Manager) LoadAll(deps ServerDependencies) error {
	paths, err := discoverPlugins(m.dir)
	if err != nil {
		m.logger.Warn("plugin directory scan failed", zap.String("dir", m.dir), zap.Error(err))
		return nil
	}
	if len(paths) == 0 {
		m.logger.Info("no plugins found", zap.String("dir", m.dir))
		return nil
	}
	for _, path := range paths {
		if err := m.loadOne(path, deps); err != nil {
			m.logger.Error("plugin load failed", zap.String("path", filepath.Base(path)), zap.Error(err))
		}
	}
	m.logger.Info("plugins loaded", zap.Int("count", len(m.plugins)))
	return nil
}

// Shutdown disables all loaded plugins in reverse order.
func (m *Manager) Shutdown() {
	for i := len(m.plugins) - 1; i >= 0; i-- {
		p := m.plugins[i]
		m.safeDisable(p)
		m.dispatcher.RemoveByOwner(p.manifest.Name)
	}
	m.plugins = nil
}

// loadOne loads and enables one plugin from a .so file.
func (m *Manager) loadOne(path string, deps ServerDependencies) error {
	factory, err := loadPluginFactory(path)
	if err != nil {
		return err
	}
	p := factory()
	manifest := p.Manifest()
	if manifest.Name == "" {
		return fmt.Errorf("plugin at %s has empty manifest name", filepath.Base(path))
	}
	for _, existing := range m.plugins {
		if existing.manifest.Name == manifest.Name {
			return fmt.Errorf("duplicate plugin name: %s", manifest.Name)
		}
	}
	srv := newServerImpl(manifest.Name, m.dispatcher, deps, m.logger)
	if err := p.Enable(srv); err != nil {
		return fmt.Errorf("plugin %s enable failed: %w", manifest.Name, err)
	}
	m.plugins = append(m.plugins, loadedPlugin{plugin: p, manifest: manifest, server: srv})
	m.logger.Info("plugin enabled", zap.String("name", manifest.Name), zap.String("version", manifest.Version))
	return nil
}

// safeDisable calls Disable on a plugin with panic recovery.
func (m *Manager) safeDisable(p loadedPlugin) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Error("plugin disable panicked", zap.String("name", p.manifest.Name), zap.Any("panic", r))
		}
	}()
	if err := p.plugin.Disable(); err != nil {
		m.logger.Warn("plugin disable error", zap.String("name", p.manifest.Name), zap.Error(err))
	}
}

// Dispatcher returns the event dispatcher for external fire calls.
func (m *Manager) Dispatcher() *Dispatcher {
	return m.dispatcher
}
