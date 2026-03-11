package users

// Config defines user module security and session settings.
type Config struct {
	// JWTSecret defines the signing key for user auth tokens.
	JWTSecret string `mapstructure:"jwt_secret"`
	// PasswordCost defines hashing cost for password storage.
	PasswordCost int `mapstructure:"password_cost" default:"12"`
	// SessionTTLSeconds defines session expiration in seconds.
	SessionTTLSeconds int `mapstructure:"session_ttl_seconds" default:"86400"`
}
