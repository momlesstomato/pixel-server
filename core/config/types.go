package config

import (
	"github.com/momlesstomato/pixel-server/core/app"
	"github.com/momlesstomato/pixel-server/core/logging"
	corepermission "github.com/momlesstomato/pixel-server/core/permission"
	"github.com/momlesstomato/pixel-server/core/postgres"
	"github.com/momlesstomato/pixel-server/core/redis"
	"github.com/momlesstomato/pixel-server/core/status"
	"github.com/momlesstomato/pixel-server/core/users"
	authenticationdomain "github.com/momlesstomato/pixel-server/pkg/authentication/domain"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
)

// Config defines the complete application configuration tree.
type Config struct {
	// App contains process-level runtime settings.
	App app.Config `mapstructure:"app"`
	// Redis contains cache and ephemeral state configuration.
	Redis redis.Config `mapstructure:"redis"`
	// PostgreSQL contains relational persistence configuration.
	PostgreSQL postgres.Config `mapstructure:"postgres"`
	// Users contains authentication and user service settings.
	Users users.Config `mapstructure:"users"`
	// Logging contains structured logger settings.
	Logging logging.Config `mapstructure:"logging"`
	// Authentication contains SSO ticket policy settings.
	Authentication authenticationdomain.Config `mapstructure:"authentication"`
	// Status contains hotel status scheduling and persistence settings.
	Status status.Config `mapstructure:"status"`
	// Permission contains permission module cache and event settings.
	Permission corepermission.Config `mapstructure:"permission"`
	// Messenger contains messenger service settings.
	Messenger messengerapplication.Config `mapstructure:"messenger"`
}

// AppConfig aliases app runtime configuration shape.
type AppConfig = app.Config

// RedisConfig aliases Redis runtime configuration shape.
type RedisConfig = redis.Config

// PostgreSQLConfig aliases PostgreSQL runtime configuration shape.
type PostgreSQLConfig = postgres.Config

// UsersConfig aliases users runtime configuration shape.
type UsersConfig = users.Config

// LoggingConfig aliases logging runtime configuration shape.
type LoggingConfig = logging.Config

// AuthenticationConfig aliases authentication runtime configuration shape.
type AuthenticationConfig = authenticationdomain.Config

// StatusConfig aliases hotel status runtime configuration shape.
type StatusConfig = status.Config

// PermissionConfig aliases permission runtime configuration shape.
type PermissionConfig = corepermission.Config

// MessengerConfig aliases messenger runtime configuration shape.
type MessengerConfig = messengerapplication.Config
