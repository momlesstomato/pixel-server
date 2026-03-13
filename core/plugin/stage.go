package plugin

import (
	"fmt"

	"go.uber.org/zap"
)

// Stage initializes the plugin system during server startup.
type Stage struct {
	// Dir defines the plugin directory path.
	Dir string
	// Logger defines the structured logger.
	Logger *zap.Logger
	// Deps defines infrastructure dependencies for plugin APIs.
	Deps ServerDependencies
}

// Initialize creates a plugin manager, loads plugins, and returns it.
func (s Stage) Initialize() (*Manager, error) {
	if s.Logger == nil {
		return nil, fmt.Errorf("logger is required for plugin stage")
	}
	dispatcher := NewDispatcher(s.Logger)
	manager, err := NewManager(dispatcher, s.Logger, s.Dir)
	if err != nil {
		return nil, err
	}
	if err := manager.LoadAll(s.Deps); err != nil {
		return nil, err
	}
	return manager, nil
}
