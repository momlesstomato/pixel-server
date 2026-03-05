package log

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	cfgpkg "pixelsv/pkg/config"
)

const (
	// FormatJSON renders logs as JSON.
	FormatJSON = "json"
	// FormatConsole renders logs in readable console output.
	FormatConsole = "console"
)

// Config defines logging settings.
type Config struct {
	// Format defines output encoding: json or console.
	Format string `mapstructure:"format" default:"console"`
	// Level defines the minimum accepted log level.
	Level string `mapstructure:"level" default:"info"`
}

// BindViper configures viper defaults and environment bindings.
func BindViper(v *viper.Viper) error {
	v.AutomaticEnv()
	if err := cfgpkg.ApplyDefaultsFromTags(v, "logging", Config{}); err != nil {
		return err
	}
	if err := v.BindEnv("logging.format", "LOG_FORMAT"); err != nil {
		return fmt.Errorf("bind LOG_FORMAT: %w", err)
	}
	if err := v.BindEnv("logging.level", "LOG_LEVEL"); err != nil {
		return fmt.Errorf("bind LOG_LEVEL: %w", err)
	}
	if value := v.GetString("LOG_FORMAT"); value != "" {
		v.Set("logging.format", value)
	}
	if value := v.GetString("LOG_LEVEL"); value != "" {
		v.Set("logging.level", value)
	}
	return nil
}

// FromViper reads logging config from viper.
func FromViper(v *viper.Viper) (Config, error) {
	cfg := Config{}
	if err := v.UnmarshalKey("logging", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal logging config: %w", err)
	}
	if err := cfgpkg.FillDefaultsFromTags(&cfg); err != nil {
		return cfg, err
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Validate checks whether Config is valid.
func (c Config) Validate() error {
	if c.Level == "" {
		return ErrEmptyLevel
	}
	if c.Format != FormatJSON && c.Format != FormatConsole {
		return fmt.Errorf("%w: %s", ErrInvalidFormat, c.Format)
	}
	if _, err := zapcore.ParseLevel(c.Level); err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	return nil
}
