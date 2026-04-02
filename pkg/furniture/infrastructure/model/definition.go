package model

import "time"

// Definition stores one item definition row in PostgreSQL.
type Definition struct {
	// ID stores stable definition identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// ItemName stores internal unique item key.
	ItemName string `gorm:"size:100;uniqueIndex;not null"`
	// PublicName stores client-visible display name.
	PublicName string `gorm:"size:100;not null;default:''"`
	// ItemType stores the item category marker.
	ItemType string `gorm:"size:1;not null;default:s"`
	// Width stores base grid width in tiles.
	Width int16 `gorm:"not null;default:1"`
	// Length stores base grid length in tiles.
	Length int16 `gorm:"not null;default:1"`
	// StackHeight stores vertical stacking height.
	StackHeight float64 `gorm:"type:numeric(6,2);not null;default:1.0"`
	// CanStack stores whether items can stack on this definition.
	CanStack bool `gorm:"not null;default:true"`
	// CanSit stores whether users can sit on this item.
	CanSit bool `gorm:"not null;default:false"`
	// IsWalkable stores whether users can walk over this item.
	IsWalkable bool `gorm:"not null;default:false"`
	// SpriteID stores client-side sprite identifier.
	SpriteID int `gorm:"not null"`
	// AllowRecycle stores whether the item is recyclable.
	AllowRecycle bool `gorm:"not null;default:true"`
	// AllowTrade stores whether the item can be traded.
	AllowTrade bool `gorm:"not null;default:true"`
	// AllowMarketplaceSell stores marketplace listing permission.
	AllowMarketplaceSell bool `gorm:"not null;default:false"`
	// AllowGift stores whether the item can be gifted.
	AllowGift bool `gorm:"not null;default:true"`
	// AllowInventoryStack stores inventory stacking permission.
	AllowInventoryStack bool `gorm:"not null;default:true"`
	// InteractionType stores the behavior handler key.
	InteractionType string `gorm:"size:50;not null;default:default"`
	// InteractionModesCount stores available interaction modes.
	InteractionModesCount int16 `gorm:"not null;default:1"`
	// EffectID stores associated avatar effect identifier.
	EffectID int `gorm:"not null;default:0"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Definition.
func (Definition) TableName() string { return "item_definitions" }
