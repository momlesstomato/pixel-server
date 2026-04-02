package domain

// ItemType categorizes furniture as floor, wall, or special variant.
type ItemType string

const (
	// ItemTypeFloor identifies a floor/standing furniture item.
	ItemTypeFloor ItemType = "s"
	// ItemTypeWall identifies a wall-mounted furniture item.
	ItemTypeWall ItemType = "i"
)

// InteractionType identifies the behavior handler for an item definition.
type InteractionType string

const (
	// InteractionDefault identifies a standard non-interactive item.
	InteractionDefault InteractionType = "default"
	// InteractionGate identifies a walkable gate item.
	InteractionGate InteractionType = "gate"
	// InteractionTeleport identifies a teleporter pair item.
	InteractionTeleport InteractionType = "teleport"
	// InteractionRoller identifies a floor roller item.
	InteractionRoller InteractionType = "roller"
	// InteractionDice identifies a throwable dice item.
	InteractionDice InteractionType = "dice"
)

// Definition defines one static furniture type metadata row.
type Definition struct {
	// ID stores stable definition identifier.
	ID int
	// ItemName stores internal unique item key.
	ItemName string
	// PublicName stores client-visible display name.
	PublicName string
	// ItemType stores the item category marker.
	ItemType ItemType
	// Width stores base grid width in tiles.
	Width int
	// Length stores base grid length in tiles.
	Length int
	// StackHeight stores vertical stacking height.
	StackHeight float64
	// CanStack stores whether items can stack on this definition.
	CanStack bool
	// CanSit stores whether users can sit on this item.
	CanSit bool
	// IsWalkable stores whether users can walk over this item.
	IsWalkable bool
	// SpriteID stores client-side sprite identifier.
	SpriteID int
	// AllowRecycle stores whether the item is recyclable.
	AllowRecycle bool
	// AllowTrade stores whether the item can be traded.
	AllowTrade bool
	// AllowMarketplaceSell stores whether marketplace listing is allowed.
	AllowMarketplaceSell bool
	// AllowGift stores whether the item can be gifted.
	AllowGift bool
	// AllowInventoryStack stores whether inventory stacking is allowed.
	AllowInventoryStack bool
	// InteractionType stores the behavior handler key.
	InteractionType InteractionType
}
