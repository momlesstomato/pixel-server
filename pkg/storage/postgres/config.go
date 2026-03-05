package postgres

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	cfgpkg "pixelsv/pkg/config"
)

// ErrEmptyURL indicates that POSTGRES_URL is required.
var ErrEmptyURL = errors.New("postgres url is required")

// Config defines PostgreSQL adapter settings.
type Config struct {
	// URL is the PostgreSQL connection URL.
	URL string `mapstructure:"url"`
	// MinConns is the minimum pool size.
	MinConns int32 `mapstructure:"min_conns" default:"1"`
	// MaxConns is the maximum pool size.
	MaxConns int32 `mapstructure:"max_conns" default:"10"`
}

// BindViper configures viper defaults and environment bindings.
func BindViper(v *viper.Viper) error {
	v.AutomaticEnv()
	if err := cfgpkg.ApplyDefaultsFromTags(v, "storage.postgres", Config{}); err != nil {
		return err
	}
	if err := v.BindEnv("storage.postgres.url", "POSTGRES_URL"); err != nil {
		return fmt.Errorf("bind POSTGRES_URL: %w", err)
	}
	if err := v.BindEnv("storage.postgres.min_conns", "POSTGRES_MIN_CONNS"); err != nil {
		return fmt.Errorf("bind POSTGRES_MIN_CONNS: %w", err)
	}
	if err := v.BindEnv("storage.postgres.max_conns", "POSTGRES_MAX_CONNS"); err != nil {
		return fmt.Errorf("bind POSTGRES_MAX_CONNS: %w", err)
	}
	if value := v.GetString("POSTGRES_URL"); value != "" {
		v.Set("storage.postgres.url", value)
	}
	if value := v.GetInt32("POSTGRES_MIN_CONNS"); value > 0 {
		v.Set("storage.postgres.min_conns", value)
	}
	if value := v.GetInt32("POSTGRES_MAX_CONNS"); value > 0 {
		v.Set("storage.postgres.max_conns", value)
	}
	return nil
}

// FromViper reads postgres config from viper.
func FromViper(v *viper.Viper) (Config, error) {
	cfg := Config{}
	if err := v.UnmarshalKey("storage.postgres", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal postgres config: %w", err)
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
	if c.URL == "" {
		return ErrEmptyURL
	}
	if c.MinConns < 0 || c.MaxConns < 1 || c.MinConns > c.MaxConns {
		return fmt.Errorf("invalid pool bounds: min=%d max=%d", c.MinConns, c.MaxConns)
	}
	return nil
}
