package domain

import "time"

// Category defines one navigator category row.
type Category struct {
	// ID stores stable navigator category identifier.
	ID int
	// Caption stores the category display name.
	Caption string
	// Visible stores whether the category is displayed.
	Visible bool
	// OrderNum stores the display sort position.
	OrderNum int
	// IconImage stores the icon index for the client.
	IconImage int
	// CategoryType stores the category classification key.
	CategoryType string
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}

// SavedSearch defines one per-user saved navigator search.
type SavedSearch struct {
	// ID stores stable saved search identifier.
	ID int
	// UserID stores the owning user identifier.
	UserID int
	// SearchCode stores the search tab key.
	SearchCode string
	// Filter stores the user filter string.
	Filter string
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
}
