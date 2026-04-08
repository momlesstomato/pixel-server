package packet

// ModKickUserPacketID defines moderator kick (c2s 2582).
const ModKickUserPacketID uint16 = 2582

// ModMuteUserPacketID defines moderator mute (c2s 1945).
const ModMuteUserPacketID uint16 = 1945

// ModBanUserPacketID defines moderator ban (c2s 2766).
const ModBanUserPacketID uint16 = 2766

// ModWarnUserPacketID defines moderator warn/caution (c2s 1840).
const ModWarnUserPacketID uint16 = 1840

// ModAlertUserPacketID defines moderator alert/message (c2s 229).
const ModAlertUserPacketID uint16 = 229

// ModRoomAlertPacketID defines moderator current-room alert (c2s 3842).
const ModRoomAlertPacketID uint16 = 3842

// SanctionTradeLockPacketID defines trade lock sanction (c2s 3742).
const SanctionTradeLockPacketID uint16 = 3742

// CallForHelpPacketID defines call-for-help submit (c2s 1691).
const CallForHelpPacketID uint16 = 1691

// GetCFHStatusPacketID defines call-for-help status query (c2s 2746).
const GetCFHStatusPacketID uint16 = 2746

// ModeratorInitPacketID defines moderator tool init (s2c 2696).
const ModeratorInitPacketID uint16 = 2696

// CFHTopicsPacketID defines call-for-help topic list (s2c 325).
const CFHTopicsPacketID uint16 = 325

// CFHPendingPacketID defines pending call-for-help list (s2c 1121).
const CFHPendingPacketID uint16 = 1121

// CFHResultPacketID defines call-for-help result (s2c 3635).
const CFHResultPacketID uint16 = 3635
