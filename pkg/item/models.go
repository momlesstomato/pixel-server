package item

// Item represents a room item (floor or wall).
type Item struct {
	ID           int32
	UserID       int32
	RoomID       int32
	BaseItem     int32
	ExtraData    string
	X            int32
	Y            int32
	Z            float64
	Rotation     int32
	WallPos      string
	LimitedNum   int32
	LimitedTotal int32
}

// Furniture describes the base definition of an item type.
type Furniture struct {
	ID                  int32
	ItemName            string
	SpriteID            int32
	PublicName          string
	Type                string // "s" floor, "i" wall
	Width               int32
	Length              int32
	Height              float64
	CanStack            bool
	CanSit              bool
	CanWalk             bool
	AllowRecycle        bool
	AllowTrade          bool
	AllowMarketplace    bool
	AllowGift           bool
	AllowInventoryStack bool
	InteractionType     string
	InteractionCount    int32
	EffectMale          int32
	EffectFemale        int32
}
