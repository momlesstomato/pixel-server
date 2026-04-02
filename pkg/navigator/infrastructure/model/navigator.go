package model

import "time"

// Category stores one navigator category row in PostgreSQL.
type Category struct {
	// ID stores stable navigator category identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// Caption stores the category display name.
	Caption string `gorm:"size:100;not null"`
	// Visible stores whether the category is displayed.
	Visible bool `gorm:"not null;default:true"`
	// OrderNum stores the display sort position.
	OrderNum int `gorm:"not null;default:0"`
	// IconImage stores the icon index for the client.
	IconImage int `gorm:"not null;default:0"`
	// CategoryType stores the category classification key.
	CategoryType string `gorm:"size:50;not null;default:public"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for Category.
func (Category) TableName() string { return "navigator_categories" }

// SavedSearch stores one per-user saved navigator search in PostgreSQL.
type SavedSearch struct {
	// ID stores stable saved search identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// UserID stores the owning user identifier.
	UserID uint `gorm:"not null;index:idx_nav_saved_user"`
	// SearchCode stores the search tab key.
	SearchCode string `gorm:"size:50;not null"`
	// Filter stores the user filter string.
	Filter string `gorm:"size:255;not null;default:''"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// TableName returns the PostgreSQL table name for SavedSearch.
func (SavedSearch) TableName() string { return "navigator_saved_searches" }
