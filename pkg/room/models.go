package room

// Room represents a guest room.
type Room struct {
	ID               int32
	OwnerID          int32
	OwnerName        string
	Name             string
	Description      string
	ModelName        string
	Password         string
	State            string // "open", "locked", "password", "invisible"
	UsersNow         int32
	UsersMax         int32
	Category         int32
	Score            int32
	PaperFloor       string
	PaperWall        string
	PaperLandscape   string
	FloorThickness   int32
	WallThickness    int32
	WallHeight       int32
	HideWall         bool
	AllowPets        bool
	AllowPetsEat     bool
	AllowWalkthrough bool
	ChatMode         int32
	ChatWeight       int32
	ChatSpeed        int32
	ChatHearRange    int32
	ChatProtection   int32
	TradeMode        int32
	RollerSpeed      int32
	MuteOption       int32
	KickOption       int32
	BanOption        int32
	Tags             string
	Group            int32
}

// Model stores the static layout of a room.
type Model struct {
	Name      string
	DoorX     int32
	DoorY     int32
	DoorZ     float64
	DoorDir   int32
	Heightmap string
	ClubOnly  bool
}
