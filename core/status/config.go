package status

// Config defines hotel status scheduling and persistence configuration.
type Config struct {
	// OpenHour defines UTC hour when hotel opens each day.
	OpenHour int `mapstructure:"open_hour" default:"0"`
	// OpenMinute defines UTC minute when hotel opens each day.
	OpenMinute int `mapstructure:"open_minute" default:"0"`
	// CloseHour defines UTC hour when hotel closes each day.
	CloseHour int `mapstructure:"close_hour" default:"23"`
	// CloseMinute defines UTC minute when hotel closes each day.
	CloseMinute int `mapstructure:"close_minute" default:"59"`
	// RedisKey defines Redis key storing serialized hotel status.
	RedisKey string `mapstructure:"redis_key" default:"hotel:status"`
	// BroadcastChannel defines broadcaster channel for global hotel packet dispatch.
	BroadcastChannel string `mapstructure:"broadcast_channel" default:"broadcast:all"`
	// CountdownTickSeconds defines ticker interval for closing countdown in seconds.
	CountdownTickSeconds int `mapstructure:"countdown_tick_seconds" default:"60"`
	// DefaultMaintenanceDurationMinutes defines default maintenance duration.
	DefaultMaintenanceDurationMinutes int `mapstructure:"default_maintenance_duration_minutes" default:"15"`
}
