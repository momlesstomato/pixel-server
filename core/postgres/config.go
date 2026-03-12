package postgres

// Config defines PostgreSQL connection, migration, and seed settings.
type Config struct {
	// DSN defines the PostgreSQL DSN connection string.
	DSN string `mapstructure:"dsn"`
	// MaxOpenConns defines the max number of open connections.
	MaxOpenConns int `mapstructure:"max_open_conns" default:"30"`
	// MaxIdleConns defines the max number of idle connections.
	MaxIdleConns int `mapstructure:"max_idle_conns" default:"10"`
	// ConnMaxLifetimeSeconds defines connection lifetime in seconds.
	ConnMaxLifetimeSeconds int `mapstructure:"conn_max_lifetime_seconds" default:"300"`
	// MigrationAutoUp runs schema migrations during initialization when enabled.
	MigrationAutoUp bool `mapstructure:"migration_auto_up" default:"false"`
	// SeedAutoUp runs essential seeders during initialization when enabled.
	SeedAutoUp bool `mapstructure:"seed_auto_up" default:"false"`
	// MigrationTable defines migration state table name.
	MigrationTable string `mapstructure:"migration_table" default:"schema_migrations"`
	// SeedTable defines seed state table name.
	SeedTable string `mapstructure:"seed_table" default:"schema_seeds"`
}
