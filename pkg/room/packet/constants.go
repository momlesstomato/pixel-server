package packet

// OpenFlatConnectionPacketID defines client room.open_flat_connection (c2s 2312).
const OpenFlatConnectionPacketID uint16 = 2312

// GetRoomEntryDataPacketID defines client room.get_room_entry_data (c2s 2300).
const GetRoomEntryDataPacketID uint16 = 2300

// MoveAvatarPacketID defines client room.move_avatar (c2s 3320).
const MoveAvatarPacketID uint16 = 3320

// ChatPacketID defines client room.chat (c2s 1314).
const ChatPacketID uint16 = 1314

// ShoutPacketID defines client room.shout (c2s 2085).
const ShoutPacketID uint16 = 2085

// WhisperPacketID defines client room.whisper (c2s 1543).
const WhisperPacketID uint16 = 1543

// DancePacketID defines client room.dance (c2s 2080).
const DancePacketID uint16 = 2080

// ActionPacketID defines client room.action (c2s 2456).
const ActionPacketID uint16 = 2456

// SitPacketID defines client room.unit_posture (c2s 2235).
const SitPacketID uint16 = 2235

// SignPacketID defines client room.sign (c2s 1975).
const SignPacketID uint16 = 1975

// StartTypingPacketID defines client room.start_typing (c2s 1597).
const StartTypingPacketID uint16 = 1597

// CancelTypingPacketID defines client room.cancel_typing (c2s 1474).
const CancelTypingPacketID uint16 = 1474

// LookToPacketID defines client room.look_to (c2s 3301).
const LookToPacketID uint16 = 3301

// LetUserInPacketID defines client room.let_user_in (c2s 1644).
const LetUserInPacketID uint16 = 1644

// KickUserPacketID defines client room.kick_user (c2s 1320).
const KickUserPacketID uint16 = 1320

// BanUserPacketID defines client room.ban_user (c2s 1477).
const BanUserPacketID uint16 = 1477

// CloseConnectionPacketID defines client flat close request (c2s 3997).
const CloseConnectionPacketID uint16 = 3997

// DesktopViewPacketID defines client hotel view request (c2s 105).
const DesktopViewPacketID uint16 = 105

// DesktopViewComposerID defines server hotel view redirect (s2c 122).
const DesktopViewComposerID uint16 = 122

// RoomReadyComposerID defines server room.ready (s2c 2031).
const RoomReadyComposerID uint16 = 2031

// OpenConnectionComposerID defines server room.open_connection (s2c 758).
const OpenConnectionComposerID uint16 = 758

// HeightMapComposerID defines server room.heightmap (s2c 2753).
const HeightMapComposerID uint16 = 2753

// FloorHeightMapComposerID defines server room.floor_heightmap (s2c 1301).
const FloorHeightMapComposerID uint16 = 1301

// RoomEntryInfoComposerID defines server room.entry_info (s2c 749).
const RoomEntryInfoComposerID uint16 = 749

// RoomVisualizationComposerID defines server room.visualization (s2c 3547).
const RoomVisualizationComposerID uint16 = 3547

// UsersComposerID defines server room.users (s2c 374).
const UsersComposerID uint16 = 374

// UserUpdateComposerID defines server room.user_update (s2c 1640).
const UserUpdateComposerID uint16 = 1640

// UserRemoveComposerID defines server room.user_remove (s2c 2661).
const UserRemoveComposerID uint16 = 2661

// ChatComposerID defines server room.chat (s2c 1446).
const ChatComposerID uint16 = 1446

// ShoutComposerID defines server room.shout (s2c 1036).
const ShoutComposerID uint16 = 1036

// WhisperComposerID defines server room.whisper (s2c 2704).
const WhisperComposerID uint16 = 2704

// DanceComposerID defines server room.dance (s2c 2233).
const DanceComposerID uint16 = 2233

// ActionComposerID defines server room.unit_expression (s2c 1631).
const ActionComposerID uint16 = 1631

// UserTypingComposerID defines server room.user_typing (s2c 1717).
const UserTypingComposerID uint16 = 1717

// SleepComposerID defines server room.sleep (s2c 1797).
const SleepComposerID uint16 = 1797

// FloodControlComposerID defines server room.flood_control (s2c 566).
const FloodControlComposerID uint16 = 566

// DoorbellComposerID defines server room.doorbell (s2c 2309).
const DoorbellComposerID uint16 = 2309

// FlatAccessibleComposerID defines server room.flat_accessible (s2c 3783).
const FlatAccessibleComposerID uint16 = 3783

// CantConnectComposerID defines server room.cant_connect (s2c 899).
const CantConnectComposerID uint16 = 899

// CloseConnectionComposerID defines server room.close_connection (s2c 122).
const CloseConnectionComposerID uint16 = 122

// FurnitureAliasesComposerID defines server room.furniture_aliases (s2c 1723).
const FurnitureAliasesComposerID uint16 = 1723

// GetRoomSettingsPacketID defines client room.get_room_settings (c2s 3129).
const GetRoomSettingsPacketID uint16 = 3129

// SaveRoomSettingsPacketID defines client room.save_room_settings (c2s 1969).
const SaveRoomSettingsPacketID uint16 = 1969

// RoomSettingsComposerID defines server room.room_settings (s2c 1498).
const RoomSettingsComposerID uint16 = 1498

// RoomSettingsSavedComposerID defines server room.room_settings_saved (s2c 948).
const RoomSettingsSavedComposerID uint16 = 948

// GiveRoomScorePacketID defines client room.give_room_score (c2s 3616).
const GiveRoomScorePacketID uint16 = 3616

// RoomScoreComposerID defines server room.room_score (s2c 3271).
const RoomScoreComposerID uint16 = 3271

// DeleteRoomPacketID defines client room.delete_room (c2s 532).
const DeleteRoomPacketID uint16 = 532

// GetBannedUsersPacketID defines client room.get_banned_users (c2s 2652).
const GetBannedUsersPacketID uint16 = 2652

// BannedUsersComposerID defines server room.banned_users (s2c 1869).
const BannedUsersComposerID uint16 = 1869

// UnbanUserPacketID defines client room.unban_user (c2s 3842).
const UnbanUserPacketID uint16 = 3842

// RoomForwardComposerID defines server room.room_forward (s2c 160).
const RoomForwardComposerID uint16 = 160

// AssignRightsPacketID defines client room.assign_rights (c2s 3843).
const AssignRightsPacketID uint16 = 3843

// RemoveRightsPacketID defines client room.remove_rights (c2s 877).
const RemoveRightsPacketID uint16 = 877

// RemoveMyRightsPacketID defines client room.remove_my_rights (c2s 111).
const RemoveMyRightsPacketID uint16 = 111

// RemoveAllRightsPacketID defines client room.remove_all_rights (c2s 884).
const RemoveAllRightsPacketID uint16 = 884

// GetRoomRightsPacketID defines client room.get_room_rights (c2s 3937).
const GetRoomRightsPacketID uint16 = 3937

// ToggleMuteToolPacketID defines client room.toggle_mute_tool (c2s 1301).
const ToggleMuteToolPacketID uint16 = 1301

// YouAreControllerComposerID defines server room.you_are_controller (s2c 680).
const YouAreControllerComposerID uint16 = 680

// RoomRightsListComposerID defines server room.rights_list (s2c 225).
const RoomRightsListComposerID uint16 = 225
