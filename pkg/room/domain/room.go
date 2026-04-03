package domain

// AccessState defines the room access control mode.
type AccessState string

const (
	// AccessOpen allows unrestricted entry.
	AccessOpen AccessState = "open"
	// AccessLocked requires doorbell approval.
	AccessLocked AccessState = "locked"
	// AccessPassword requires a password match.
	AccessPassword AccessState = "password"
	// AccessInvisible hides the room from navigator.
	AccessInvisible AccessState = "invisible"
)

// Room defines the persistent room aggregate.
type Room struct {
	// ID stores the stable room identifier.
	ID int
	// OwnerID stores the room creator identifier.
	OwnerID int
	// OwnerName stores the room creator display name.
	OwnerName string
	// Name stores the room display name.
	Name string
	// Description stores the room description text.
	Description string
	// State stores the room access state.
	State AccessState
	// ModelSlug stores the room model template identifier.
	ModelSlug string
	// CustomHeightmap stores the user-created heightmap override.
	CustomHeightmap string
	// CategoryID stores the navigator category reference.
	CategoryID int
	// MaxUsers stores the room capacity limit.
	MaxUsers int
	// Password stores the bcrypt password hash for password rooms.
	Password string
	// WallHeight stores the custom wall height (-1 = auto).
	WallHeight int
	// FloorThickness stores the floor rendering thickness.
	FloorThickness int
	// WallThickness stores the wall rendering thickness.
	WallThickness int
	// AllowPets reports whether pet placement is allowed.
	AllowPets bool
	// AllowTrading reports whether trading is enabled.
	AllowTrading bool
	// Score stores the room star rating.
	Score int
	// Tags stores the room searchable tags.
	Tags []string
	// TradeMode stores the trade policy code.
	TradeMode int
}

// RoomModel defines a predefined room template layout.
type RoomModel struct {
	// ID stores the stable model identifier.
	ID int
	// Slug stores the model string identifier.
	Slug string
	// Heightmap stores the raw heightmap string.
	Heightmap string
	// DoorX stores the door tile horizontal coordinate.
	DoorX int
	// DoorY stores the door tile vertical coordinate.
	DoorY int
	// DoorZ stores the door tile height.
	DoorZ float64
	// DoorDir stores the door facing direction (0-7).
	DoorDir int
	// WallHeight stores the custom wall height (-1 for auto).
	WallHeight int
}
