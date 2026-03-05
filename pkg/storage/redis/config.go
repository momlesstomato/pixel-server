package redis

import (
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	cfgpkg "pixelsv/pkg/config"
)

// ErrEmptyURL indicates that REDIS_URL is required.
var ErrEmptyURL = errors.New("redis url is required")

// Config defines Redis adapter settings.
type Config struct {
	// URL is the Redis connection URL.
	URL string `mapstructure:"url"`
	// KeyPrefix prefixes all adapter-owned keys.
	KeyPrefix string `mapstructure:"key_prefix" default:"pixelsv"`
	// SessionTTLSeconds controls session persistence TTL.
	SessionTTLSeconds int `mapstructure:"session_ttl_seconds" default:"3600"`
}

// BindViper configures viper defaults and environment bindings.
func BindViper(v *viper.Viper) error {
	v.AutomaticEnv()
	if err := cfgpkg.ApplyDefaultsFromTags(v, "storage.redis", Config{}); err != nil {
		return err
	}
	if err := v.BindEnv("storage.redis.url", "REDIS_URL"); err != nil {
		return fmt.Errorf("bind REDIS_URL: %w", err)
	}
	if err := v.BindEnv("storage.redis.key_prefix", "REDIS_KEY_PREFIX"); err != nil {
		return fmt.Errorf("bind REDIS_KEY_PREFIX: %w", err)
	}
	if err := v.BindEnv("storage.redis.session_ttl_seconds", "REDIS_SESSION_TTL_SECONDS"); err != nil {
		return fmt.Errorf("bind REDIS_SESSION_TTL_SECONDS: %w", err)
	}
	if value := v.GetString("REDIS_URL"); value != "" {
		v.Set("storage.redis.url", value)
	}
	if value := v.GetString("REDIS_KEY_PREFIX"); value != "" {
		v.Set("storage.redis.key_prefix", value)
	}
	if value := v.GetInt("REDIS_SESSION_TTL_SECONDS"); value > 0 {
		v.Set("storage.redis.session_ttl_seconds", value)
	}
	return nil
}

// FromViper reads redis config from viper.
func FromViper(v *viper.Viper) (Config, error) {
	cfg := Config{}
	if err := v.UnmarshalKey("storage.redis", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal redis config: %w", err)
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
	if _, err := redis.ParseURL(c.URL); err != nil {
		return fmt.Errorf("parse redis url: %w", err)
	}
	if c.SessionTTLSeconds < 1 {
		return fmt.Errorf("invalid session ttl: %d", c.SessionTTLSeconds)
	}
	return nil
}
