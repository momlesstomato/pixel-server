package logging

import (
	"fmt"
	"io"

	"github.com/momlesstomato/pixel-server/core/config"
	"go.uber.org/zap"
)

// Stage defines logging startup behavior.
type Stage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializeLogger creates a logger from loaded configuration.
	InitializeLogger(*config.Config) (*zap.Logger, error)
}

// Initializer provides default logger startup behavior.
type Initializer struct {
	// Output defines the destination stream for logs.
	Output io.Writer
}

// Name returns the stable initializer name.
func (initializer Initializer) Name() string {
	return "logger"
}

// InitializeLogger builds and returns a configured logger.
func (initializer Initializer) InitializeLogger(loaded *config.Config) (*zap.Logger, error) {
	if loaded == nil {
		return nil, fmt.Errorf("config is required")
	}
	return New(loaded.Logging, initializer.Output)
}
