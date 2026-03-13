package model

// WardrobeSlot stores one persisted user wardrobe slot row.
type WardrobeSlot struct {
	// ID stores stable slot row identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores owning user identifier.
	UserID uint `gorm:"not null;uniqueIndex:idx_user_wardrobe_user_slot,priority:1"`
	// SlotID stores wardrobe slot index.
	SlotID int `gorm:"not null;uniqueIndex:idx_user_wardrobe_user_slot,priority:2"`
	// Figure stores saved avatar figure string.
	Figure string `gorm:"size:255;not null"`
	// Gender stores saved avatar gender marker.
	Gender string `gorm:"size:1;not null;default:M"`
}
