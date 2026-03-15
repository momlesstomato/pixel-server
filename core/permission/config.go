package permission

// Config defines permission module runtime configuration.
type Config struct {
	// CachePrefix defines Redis key prefix for group snapshots.
	CachePrefix string `mapstructure:"cache_prefix" default:"perm:group"`
	// CacheTTLSeconds defines Redis cache TTL in seconds.
	CacheTTLSeconds int `mapstructure:"cache_ttl_seconds" default:"300"`
	// AmbassadorPermission defines dotted permission granting ambassador flag.
	AmbassadorPermission string `mapstructure:"ambassador_permission" default:"role.ambassador"`
	// EmitPermissionChecked defines whether plugin permission checks fire events.
	EmitPermissionChecked bool `mapstructure:"emit_permission_checked" default:"false"`
}
