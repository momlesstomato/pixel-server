package domain

const (
	// PermTool allows access to the moderation tool interface on login.
	PermTool = "moderation.tool"
	// PermKick allows issuing a hotel-level kick (force disconnect).
	PermKick = "moderation.kick"
	// PermBan allows issuing a hotel-level ban (account/IP/machine).
	PermBan = "moderation.ban"
	// PermMute allows issuing a hotel-level mute (chat restriction).
	PermMute = "moderation.mute"
	// PermWarn allows sending a warning or caution notice to a user.
	PermWarn = "moderation.warn"
	// PermTradeLock allows issuing a hotel-level trade lock sanction.
	PermTradeLock = "moderation.trade_lock"
	// PermUnban allows deactivating an active hotel ban.
	PermUnban = "moderation.unban"
	// PermUnmute allows deactivating an active hotel mute.
	PermUnmute = "moderation.unmute"
	// PermHistory allows viewing a user's moderation action history.
	PermHistory = "moderation.history"
	// PermAmbassador identifies the ambassador role for alert broadcasts.
	PermAmbassador = "role.ambassador"
)
