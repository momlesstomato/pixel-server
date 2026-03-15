package model

// Grant defines one persisted group_permissions row.
type Grant struct {
	// GroupID stores owning group identifier.
	GroupID uint `gorm:"primaryKey;not null"`
	// Permission stores dotted-notation permission string.
	Permission string `gorm:"primaryKey;size:128;not null;index"`
}
