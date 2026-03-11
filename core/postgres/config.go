package postgres

// Config defines PostgreSQL connection and pool settings.
type Config struct {
	// DSN defines the PostgreSQL DSN connection string.
	DSN string `mapstructure:"dsn"`
	// MaxOpenConns defines the max number of open connections.
	MaxOpenConns int `mapstructure:"max_open_conns" default:"30"`
	// MaxIdleConns defines the max number of idle connections.
	MaxIdleConns int `mapstructure:"max_idle_conns" default:"10"`
	// ConnMaxLifetimeSeconds defines connection lifetime in seconds.
	ConnMaxLifetimeSeconds int `mapstructure:"conn_max_lifetime_seconds" default:"300"`
}
