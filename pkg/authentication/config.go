package authentication

// Config defines SSO token issuance policy settings.
type Config struct {
	// DefaultTTLSeconds defines fallback token lifetime in seconds.
	DefaultTTLSeconds int `mapstructure:"default_ttl_seconds" default:"300"`
	// MaxTTLSeconds defines maximum allowed token lifetime in seconds.
	MaxTTLSeconds int `mapstructure:"max_ttl_seconds" default:"1800"`
	// KeyPrefix defines Redis key namespace for issued tickets.
	KeyPrefix string `mapstructure:"key_prefix" default:"sso"`
}
