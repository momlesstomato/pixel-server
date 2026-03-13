package domain

// DefaultDailyRespects defines max user-to-user respects allowed per UTC day.
const DefaultDailyRespects = 3

// Settings defines one user client settings value object.
type Settings struct {
	// UserID stores owning user identifier.
	UserID int
	// VolumeSystem stores global system volume percentage.
	VolumeSystem int
	// VolumeFurni stores furniture volume percentage.
	VolumeFurni int
	// VolumeTrax stores trax volume percentage.
	VolumeTrax int
	// OldChat stores classic chat style preference.
	OldChat bool
	// RoomInvites stores room invite preference.
	RoomInvites bool
	// CameraFollow stores camera follow preference.
	CameraFollow bool
	// Flags stores bitmask settings field.
	Flags int
	// ChatType stores client chat rendering type.
	ChatType int
}

// SettingsPatch defines partial settings update payload.
type SettingsPatch struct {
	// VolumeSystem stores optional global system volume percentage.
	VolumeSystem *int
	// VolumeFurni stores optional furniture volume percentage.
	VolumeFurni *int
	// VolumeTrax stores optional trax volume percentage.
	VolumeTrax *int
	// OldChat stores optional classic chat style preference.
	OldChat *bool
	// RoomInvites stores optional room invite preference.
	RoomInvites *bool
	// CameraFollow stores optional camera follow preference.
	CameraFollow *bool
	// Flags stores optional bitmask settings field.
	Flags *int
	// ChatType stores optional client chat rendering type.
	ChatType *int
}
