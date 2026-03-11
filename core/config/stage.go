package config

// Stage defines configuration startup behavior.
type Stage interface {
	// Name returns a stable startup unit identifier.
	Name() string
	// InitializeConfig creates the application configuration.
	InitializeConfig() (*Config, error)
}

// Initializer provides default configuration startup behavior.
type Initializer struct {
	// Options defines config loading behavior.
	Options LoaderOptions
}

// Name returns the stable initializer name.
func (initializer Initializer) Name() string {
	return "config"
}

// InitializeConfig loads and returns application configuration.
func (initializer Initializer) InitializeConfig() (*Config, error) {
	return Load(initializer.Options)
}
