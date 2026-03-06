package httpserver

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
	cfgpkg "pixelsv/pkg/config"
)

// ErrEmptyAPIKey indicates API key is required for admin endpoints.
var ErrEmptyAPIKey = errors.New("api key is required")

// Config defines HTTP server settings.
type Config struct {
	// Address is the bind address.
	Address string `mapstructure:"address" default:":8080"`
	// DisableStartupMessage controls Fiber startup banner visibility.
	DisableStartupMessage bool `mapstructure:"disable_startup_message" default:"true"`
	// ReadTimeoutSeconds sets read timeout in seconds.
	ReadTimeoutSeconds int `mapstructure:"read_timeout_seconds" default:"10"`
	// OpenAPIPath sets the OpenAPI JSON endpoint path.
	OpenAPIPath string `mapstructure:"openapi_path" default:"/openapi.json"`
	// SwaggerPath sets the Swagger UI endpoint path.
	SwaggerPath string `mapstructure:"swagger_path" default:"/swagger"`
	// APIKey protects administrative endpoints.
	APIKey string `mapstructure:"api_key"`
}

// BindViper configures defaults and env bindings.
func BindViper(v *viper.Viper) error {
	v.AutomaticEnv()
	if err := cfgpkg.ApplyDefaultsFromTags(v, "http", Config{}); err != nil {
		return err
	}
	if err := v.BindEnv("http.address", "HTTP_ADDR"); err != nil {
		return fmt.Errorf("bind HTTP_ADDR: %w", err)
	}
	if err := v.BindEnv("http.api_key", "API_KEY"); err != nil {
		return fmt.Errorf("bind API_KEY: %w", err)
	}
	if err := v.BindEnv("http.openapi_path", "OPENAPI_PATH"); err != nil {
		return fmt.Errorf("bind OPENAPI_PATH: %w", err)
	}
	if err := v.BindEnv("http.swagger_path", "SWAGGER_PATH"); err != nil {
		return fmt.Errorf("bind SWAGGER_PATH: %w", err)
	}
	if err := v.BindEnv("http.read_timeout_seconds", "HTTP_READ_TIMEOUT_SECONDS"); err != nil {
		return fmt.Errorf("bind HTTP_READ_TIMEOUT_SECONDS: %w", err)
	}
	setBoundValue(v, "http.address", "HTTP_ADDR")
	setBoundValue(v, "http.api_key", "API_KEY")
	setBoundValue(v, "http.openapi_path", "OPENAPI_PATH")
	setBoundValue(v, "http.swagger_path", "SWAGGER_PATH")
	setBoundValue(v, "http.read_timeout_seconds", "HTTP_READ_TIMEOUT_SECONDS")
	return nil
}

// FromViper reads HTTP server config from viper.
func FromViper(v *viper.Viper) (Config, error) {
	cfg := Config{}
	if err := v.UnmarshalKey("http", &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal http config: %w", err)
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
	if c.APIKey == "" {
		return ErrEmptyAPIKey
	}
	if c.ReadTimeoutSeconds < 1 {
		return fmt.Errorf("invalid read timeout: %d", c.ReadTimeoutSeconds)
	}
	return nil
}

func setBoundValue(v *viper.Viper, key string, env string) {
	if value := v.Get(env); value != nil && value != "" {
		v.Set(key, value)
	}
}
