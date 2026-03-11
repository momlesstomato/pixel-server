package config

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/spf13/viper"
)

// LoaderOptions defines runtime options for configuration loading.
type LoaderOptions struct {
	// EnvFile is the .env file path; defaults to .env when empty.
	EnvFile string
	// EnvPrefix is an optional environment variable prefix.
	EnvPrefix string
}

// Load parses and validates application configuration from file and environment.
func Load(options LoaderOptions) (*Config, error) {
	instance := viper.New()
	envFile := options.EnvFile
	if envFile == "" {
		envFile = ".env"
	}
	instance.SetConfigType("env")
	instance.SetConfigFile(envFile)
	if err := instance.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read configuration file %q: %w", envFile, err)
		}
	}
	if options.EnvPrefix != "" {
		instance.SetEnvPrefix(options.EnvPrefix)
	}
	instance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	instance.AutomaticEnv()
	keys, required, err := bindDefaultsAndEnv(instance, reflect.TypeOf(Config{}), "")
	if err != nil {
		return nil, err
	}
	applyAliasValues(instance, keys, options.EnvPrefix)
	missing := missingRequiredKeys(instance, required, options.EnvPrefix)
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("missing mandatory configuration: %s", strings.Join(missing, ", "))
	}
	var cfg Config
	if err := instance.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal configuration: %w", err)
	}
	return &cfg, nil
}
