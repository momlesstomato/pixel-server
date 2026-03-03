package user

import "time"

// User represents a player account.
type User struct {
	ID               int32
	Username         string
	Password         string
	Email            string
	Motto            string
	Figure           string
	Credits          int32
	Pixels           int32
	Points           int32
	Rank             int32
	AccountCreated   time.Time
	LastLogin        time.Time
	LastOnline       time.Time
	Online           bool
	IPRegister       string
	IPCurrent        string
	HomeRoom         int32
	Gender           string // "M" or "F"
	AchievementScore int32
}

// Settings stores per-user preference flags.
type Settings struct {
	UserID                     int32
	AchievementTrackingEnabled bool
	AllowFriendRequests        bool
	AllowMimic                 bool
	AllowTrade                 bool
	AllowWhispers              bool
	BlockCameraFollow          bool
	BlockAlerts                bool
	BubbleID                   int32
	ChatColor                  int32
	DisableWhisper             bool
	ForceOpenWidget            bool
	FocusPreference            int32
	FriendBarState             int32
	GuildID                    int32
	HabboClubExpiry            time.Time
	HomeRoom                   int32
	IgnoreInvites              bool
	IgnoreRoomInvites          bool
	MentorLevel                int32
	MuteState                  bool
	NavigatorHeight            int32
	NavigatorWidth             int32
	NavigatorSearches          string
	NavigatorX                 int32
	NavigatorY                 int32
	OldChat                    bool
	OnlineTimeSeconds          int64
	RoomCameraFollow           bool
	Tags                       string
	Timestamp                  time.Time
	Volume                     string
	VolumeSystem               int32
}

// WardrobeOutfit represents a saved outfit in a wardrobe slot.
type WardrobeOutfit struct {
	UserID int32
	SlotID int32
	Look   string
	Gender string
}

// Badge represents a collectible badge owned by a user.
type Badge struct {
	UserID   int32
	Code     string
	Slot     int32 // 0 = unequipped, 1-5 = equipped slot
	Acquired int64 // unix timestamp
}

// IgnoredUser represents an entry in a user's ignore list.
type IgnoredUser struct {
	UserID        int32
	IgnoredUserID int32
	Username      string
}
