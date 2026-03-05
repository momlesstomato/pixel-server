package config

import "fmt"

// Config defines the root configuration for shared runtime concerns.
type Config struct {
	// App contains generic application settings.
	App AppConfig `mapstructure:"app"`
}

// AppConfig contains base application settings.
type AppConfig struct {
	// Env identifies the runtime environment.
	Env string `mapstructure:"env" default:"development"`
}

// Validate checks whether Config is internally consistent.
func (c Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	return nil
}

// Validate checks whether AppConfig is internally consistent.
func (c AppConfig) Validate() error {
	if c.Env == "" {
		return ErrEmptyAppEnv
	}
	return nil
}
