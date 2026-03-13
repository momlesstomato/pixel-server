package users

// Config defines user module security and session settings.
type Config struct {
	// JWTSecret defines the signing key for user auth tokens.
	JWTSecret string `mapstructure:"jwt_secret"`
	// PasswordCost defines hashing cost for password storage.
	PasswordCost int `mapstructure:"password_cost" default:"12"`
	// SessionTTLSeconds defines session expiration in seconds.
	SessionTTLSeconds int `mapstructure:"session_ttl_seconds" default:"86400"`
	// SettingsRoomInvitesPacketID defines packet mapping for user.settings_room_invites.
	SettingsRoomInvitesPacketID int `mapstructure:"settings_room_invites_packet_id" default:"65534"`
	// SettingsOldChatPacketID defines packet mapping for user.settings_old_chat.
	SettingsOldChatPacketID int `mapstructure:"settings_old_chat_packet_id" default:"65535"`
	// UnignorePacketID defines packet mapping for user.unignore.
	UnignorePacketID int `mapstructure:"unignore_packet_id" default:"65533"`
	// IgnoreByIDPacketID defines packet mapping for user.ignore_id.
	IgnoreByIDPacketID int `mapstructure:"ignore_by_id_packet_id" default:"65532"`
	// ApproveNamePacketID defines packet mapping for user.approve_name.
	ApproveNamePacketID int `mapstructure:"approve_name_packet_id" default:"65531"`
}
