package domain

import "time"

// CatalogPage defines one catalog tree node with layout and visibility.
type CatalogPage struct {
	// ID stores stable page identifier.
	ID int
	// ParentID stores parent page reference for tree hierarchy.
	ParentID *int
	// Caption stores the page display title.
	Caption string
	// IconImage stores the page icon sprite identifier.
	IconImage int
	// PageLayout stores the client layout template name.
	PageLayout string
	// Visible stores whether the page is shown in catalog.
	Visible bool
	// Enabled stores whether purchasing from this page is allowed.
	Enabled bool
	// MinPermission stores dotted permission required to view the page; empty means everyone.
	MinPermission string
	// ClubOnly stores whether only club members can access.
	ClubOnly bool
	// OrderNum stores the display ordering position.
	OrderNum int
	// Images stores page header image identifiers.
	Images []string
	// Texts stores page body text blocks.
	Texts []string
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}
