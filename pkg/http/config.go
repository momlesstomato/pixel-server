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
	// WebSocketPingIntervalSeconds sets server ping interval in seconds.
	WebSocketPingIntervalSeconds int `mapstructure:"websocket_ping_interval_seconds" default:"30"`
	// WebSocketPongTimeoutSeconds sets max pong silence timeout in seconds.
	WebSocketPongTimeoutSeconds int `mapstructure:"websocket_pong_timeout_seconds" default:"90"`
	// WebSocketAvailabilityOpen sets availability.status isOpen field.
	WebSocketAvailabilityOpen bool `mapstructure:"websocket_availability_open" default:"true"`
	// WebSocketAvailabilityOnShutdown sets availability.status onShutdown field.
	WebSocketAvailabilityOnShutdown bool `mapstructure:"websocket_availability_on_shutdown" default:"false"`
	// WebSocketAvailabilityAuthentic sets availability.status isAuthentic field.
	WebSocketAvailabilityAuthentic bool `mapstructure:"websocket_availability_authentic" default:"true"`
	// WebSocketTelemetryMinIntervalMS sets telemetry log throttle interval in milliseconds.
	WebSocketTelemetryMinIntervalMS int `mapstructure:"websocket_telemetry_min_interval_ms" default:"1000"`
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
	if err := v.BindEnv("http.websocket_ping_interval_seconds", "WS_PING_INTERVAL_SECONDS"); err != nil {
		return fmt.Errorf("bind WS_PING_INTERVAL_SECONDS: %w", err)
	}
	if err := v.BindEnv("http.websocket_pong_timeout_seconds", "WS_PONG_TIMEOUT_SECONDS"); err != nil {
		return fmt.Errorf("bind WS_PONG_TIMEOUT_SECONDS: %w", err)
	}
	if err := v.BindEnv("http.websocket_availability_open", "WS_AVAILABILITY_OPEN"); err != nil {
		return fmt.Errorf("bind WS_AVAILABILITY_OPEN: %w", err)
	}
	if err := v.BindEnv("http.websocket_availability_on_shutdown", "WS_AVAILABILITY_ON_SHUTDOWN"); err != nil {
		return fmt.Errorf("bind WS_AVAILABILITY_ON_SHUTDOWN: %w", err)
	}
	if err := v.BindEnv("http.websocket_availability_authentic", "WS_AVAILABILITY_AUTHENTIC"); err != nil {
		return fmt.Errorf("bind WS_AVAILABILITY_AUTHENTIC: %w", err)
	}
	if err := v.BindEnv("http.websocket_telemetry_min_interval_ms", "WS_TELEMETRY_MIN_INTERVAL_MS"); err != nil {
		return fmt.Errorf("bind WS_TELEMETRY_MIN_INTERVAL_MS: %w", err)
	}
	setBoundValue(v, "http.address", "HTTP_ADDR")
	setBoundValue(v, "http.api_key", "API_KEY")
	setBoundValue(v, "http.openapi_path", "OPENAPI_PATH")
	setBoundValue(v, "http.swagger_path", "SWAGGER_PATH")
	setBoundValue(v, "http.read_timeout_seconds", "HTTP_READ_TIMEOUT_SECONDS")
	setBoundValue(v, "http.websocket_ping_interval_seconds", "WS_PING_INTERVAL_SECONDS")
	setBoundValue(v, "http.websocket_pong_timeout_seconds", "WS_PONG_TIMEOUT_SECONDS")
	setBoundValue(v, "http.websocket_availability_open", "WS_AVAILABILITY_OPEN")
	setBoundValue(v, "http.websocket_availability_on_shutdown", "WS_AVAILABILITY_ON_SHUTDOWN")
	setBoundValue(v, "http.websocket_availability_authentic", "WS_AVAILABILITY_AUTHENTIC")
	setBoundValue(v, "http.websocket_telemetry_min_interval_ms", "WS_TELEMETRY_MIN_INTERVAL_MS")
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
	if c.WebSocketPingIntervalSeconds < 1 {
		return fmt.Errorf("invalid websocket ping interval: %d", c.WebSocketPingIntervalSeconds)
	}
	if c.WebSocketPongTimeoutSeconds < c.WebSocketPingIntervalSeconds {
		return fmt.Errorf("invalid websocket pong timeout: %d", c.WebSocketPongTimeoutSeconds)
	}
	if c.WebSocketTelemetryMinIntervalMS < 0 {
		return fmt.Errorf("invalid websocket telemetry min interval: %d", c.WebSocketTelemetryMinIntervalMS)
	}
	return nil
}

func setBoundValue(v *viper.Viper, key string, env string) {
	if value := v.Get(env); value != nil && value != "" {
		v.Set(key, value)
	}
}
