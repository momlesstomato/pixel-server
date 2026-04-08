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

// RoomAmbassadorAlertPacketID defines targeted ambassador alert (c2s 2996).
const RoomAmbassadorAlertPacketID uint16 = 2996

// ModToolRequestRoomInfoPacketID defines moderator room info request (c2s 707).
const ModToolRequestRoomInfoPacketID uint16 = 707

// ModToolChangeRoomSettingsPacketID defines moderator room settings action request (c2s 3260).
const ModToolChangeRoomSettingsPacketID uint16 = 3260

// ModToolRequestRoomChatlogPacketID defines moderator room chatlog request (c2s 2587).
const ModToolRequestRoomChatlogPacketID uint16 = 2587

// ModToolUserInfoPacketID defines moderator user info request (c2s 3295).
const ModToolUserInfoPacketID uint16 = 3295

// GetPendingCallsForHelpPacketID defines pending CFH request (c2s 3267).
const GetPendingCallsForHelpPacketID uint16 = 3267

// GetCFHChatlogPacketID defines CFH chatlog request (c2s 211).
const GetCFHChatlogPacketID uint16 = 211

// ModToolPreferencesPacketID defines moderator tool preferences request (c2s 31).
const ModToolPreferencesPacketID uint16 = 31

// RoomMutePacketID defines moderator room-wide mute toggle (c2s 3637).
const RoomMutePacketID uint16 = 3637

// SanctionTradeLockPacketID defines trade lock sanction (c2s 3742).
const SanctionTradeLockPacketID uint16 = 3742

// CallForHelpPacketID defines call-for-help submit (c2s 1691).
const CallForHelpPacketID uint16 = 1691

// GetCFHStatusPacketID defines call-for-help status query (c2s 2746).
const GetCFHStatusPacketID uint16 = 2746

// GuideSessionCreatePacketID defines guide-assistance session creation (c2s 3338).
const GuideSessionCreatePacketID uint16 = 3338

// GetGuideReportingStatusPacketID defines guide/reporting status query (c2s 3786).
const GetGuideReportingStatusPacketID uint16 = 3786

// ModeratorInitPacketID defines moderator tool init (s2c 2696).
const ModeratorInitPacketID uint16 = 2696

// CFHTopicsPacketID defines call-for-help topic list (s2c 325).
const CFHTopicsPacketID uint16 = 325

// CFHPendingPacketID defines pending call-for-help list (s2c 1121).
const CFHPendingPacketID uint16 = 1121

// CFHResultPacketID defines call-for-help result (s2c 3635).
const CFHResultPacketID uint16 = 3635

// CFHSanctionStatusPacketID defines call-for-help sanction status payload (s2c 2221).
const CFHSanctionStatusPacketID uint16 = 2221

// GuideSessionErrorPacketID defines guide session error payload (s2c 673).
const GuideSessionErrorPacketID uint16 = 673

// GuideReportingStatusPacketID defines guide/reporting status payload (s2c 3463).
const GuideReportingStatusPacketID uint16 = 3463

// ModToolRoomInfoComposerID defines moderator room info payload (s2c 1333).
const ModToolRoomInfoComposerID uint16 = 1333

// ModToolRoomChatlogComposerID defines moderator room chatlog payload (s2c 3434).
const ModToolRoomChatlogComposerID uint16 = 3434

// ModeratorUserInfoComposerID defines moderator user info payload (s2c 2866).
const ModeratorUserInfoComposerID uint16 = 2866

// ModeratorCFHChatlogPacketID defines moderator CFH chatlog payload (s2c 607).
const ModeratorCFHChatlogPacketID uint16 = 607

// ModeratorToolPreferencesComposerID defines moderator tool preferences payload (s2c 1576).
const ModeratorToolPreferencesComposerID uint16 = 1576

// RoomMutedComposerID defines room mute state payload (s2c 2533).
const RoomMutedComposerID uint16 = 2533
