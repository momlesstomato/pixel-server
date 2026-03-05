package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// LoadOptions defines configuration loading options.
type LoadOptions struct {
	// EnvFile points to an env file. Empty means ".env".
	EnvFile string `default:".env"`
}

// DefaultLoadOptions returns default configuration loading options.
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{EnvFile: ".env"}
}

// Load reads configuration from env file and environment variables.
func Load(opts LoadOptions) (Config, error) {
	v, err := NewViper(opts)
	if err != nil {
		return Config{}, err
	}
	return FromViper(v)
}

// NewViper configures viper defaults and environment bindings.
func NewViper(opts LoadOptions) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := BindViper(v); err != nil {
		return nil, err
	}
	if err := readEnvFile(v, opts.EnvFile); err != nil {
		return nil, err
	}
	if err := BindViper(v); err != nil {
		return nil, err
	}
	return v, nil
}

// BindViper configures defaults and environment bindings.
func BindViper(v *viper.Viper) error {
	v.AutomaticEnv()
	if err := ApplyDefaultsFromTags(v, "", Config{}); err != nil {
		return err
	}
	if err := v.BindEnv("app.env", "APP_ENV"); err != nil {
		return fmt.Errorf("bind APP_ENV: %w", err)
	}
	if value := v.GetString("APP_ENV"); value != "" {
		v.Set("app.env", value)
	}
	return nil
}

// FromViper reads shared config from viper.
func FromViper(v *viper.Viper) (Config, error) {
	cfg := Config{}
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := FillDefaultsFromTags(&cfg); err != nil {
		return cfg, err
	}
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// readEnvFile loads the env file when it exists.
func readEnvFile(v *viper.Viper, envFile string) error {
	file := envFile
	if file == "" {
		file = ".env"
	}
	if _, err := os.Stat(file); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("stat env file: %w", err)
	}
	v.SetConfigFile(file)
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("read env file: %w", err)
	}
	return nil
}
