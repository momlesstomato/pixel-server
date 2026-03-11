package config

// Config defines the complete application configuration tree.
type Config struct {
	// App contains process-level runtime settings.
	App AppConfig `mapstructure:"app"`
	// Redis contains cache and ephemeral state configuration.
	Redis RedisConfig `mapstructure:"redis"`
	// PostgreSQL contains relational persistence configuration.
	PostgreSQL PostgreSQLConfig `mapstructure:"postgres"`
	// Users contains authentication and user service settings.
	Users UsersConfig `mapstructure:"users"`
	// Logging contains structured logger settings.
	Logging LoggingConfig `mapstructure:"logging"`
}

// AppConfig defines application network and identity settings.
type AppConfig struct {
	// BindIP sets the interface address for the server listener.
	BindIP string `mapstructure:"bind_ip" default:"0.0.0.0"`
	// Port sets the network port for the server listener.
	Port int `mapstructure:"port" default:"3000"`
	// Name sets the logical service name.
	Name string `mapstructure:"name" default:"pixel-server"`
	// Environment sets the runtime environment name.
	Environment string `mapstructure:"environment" default:"development"`
}

// RedisConfig defines Redis connectivity and pool settings.
type RedisConfig struct {
	// Address defines the Redis endpoint in host:port format.
	Address string `mapstructure:"address"`
	// Password defines the Redis authentication password.
	Password string `mapstructure:"password" default:""`
	// DB defines the Redis logical database index.
	DB int `mapstructure:"db" default:"0"`
	// PoolSize defines the maximum number of pooled connections.
	PoolSize int `mapstructure:"pool_size" default:"20"`
}

// PostgreSQLConfig defines PostgreSQL connection and pool settings.
type PostgreSQLConfig struct {
	// DSN defines the PostgreSQL DSN connection string.
	DSN string `mapstructure:"dsn"`
	// MaxOpenConns defines the max number of open connections.
	MaxOpenConns int `mapstructure:"max_open_conns" default:"30"`
	// MaxIdleConns defines the max number of idle connections.
	MaxIdleConns int `mapstructure:"max_idle_conns" default:"10"`
	// ConnMaxLifetimeSeconds defines connection lifetime in seconds.
	ConnMaxLifetimeSeconds int `mapstructure:"conn_max_lifetime_seconds" default:"300"`
}

// UsersConfig defines user module security and session settings.
type UsersConfig struct {
	// JWTSecret defines the signing key for user auth tokens.
	JWTSecret string `mapstructure:"jwt_secret"`
	// PasswordCost defines hashing cost for password storage.
	PasswordCost int `mapstructure:"password_cost" default:"12"`
	// SessionTTLSeconds defines session expiration in seconds.
	SessionTTLSeconds int `mapstructure:"session_ttl_seconds" default:"86400"`
}

// LoggingConfig defines logger output and verbosity behavior.
type LoggingConfig struct {
	// Format selects structured output format: json or console.
	Format string `mapstructure:"format" default:"console"`
	// Level selects threshold level: debug, info, warn, or error.
	Level string `mapstructure:"level" default:"info"`
}
