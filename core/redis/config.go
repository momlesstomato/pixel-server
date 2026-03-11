package redis

// Config defines Redis connectivity and pool settings.
type Config struct {
	// Address defines the Redis endpoint in host:port format.
	Address string `mapstructure:"address"`
	// Password defines the Redis authentication password.
	Password string `mapstructure:"password" default:""`
	// DB defines the Redis logical database index.
	DB int `mapstructure:"db" default:"0"`
	// PoolSize defines the maximum number of pooled connections.
	PoolSize int `mapstructure:"pool_size" default:"20"`
}
