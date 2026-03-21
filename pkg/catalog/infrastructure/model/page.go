package model

import "time"

// Page stores one catalog page row in PostgreSQL.
type Page struct {
	// ID stores stable page identifier.
	ID uint `gorm:"primaryKey;autoIncrement"`
	// ParentID stores parent page reference for tree hierarchy.
	ParentID *uint `gorm:"index:idx_catalog_pages_parent"`
	// Caption stores the page display title.
	Caption string `gorm:"size:100;not null"`
	// IconImage stores the page icon sprite identifier.
	IconImage int `gorm:"not null;default:0"`
	// PageLayout stores the client layout template name.
	PageLayout string `gorm:"size:50;not null;default:default_3x3"`
	// Visible stores whether the page is shown in catalog.
	Visible bool `gorm:"not null;default:true;index:idx_catalog_pages_visible"`
	// Enabled stores whether purchasing from this page is allowed.
	Enabled bool `gorm:"not null;default:true"`
	// MinRank stores minimum rank required to view the page.
	MinRank int `gorm:"not null;default:1"`
	// ClubOnly stores whether only club members can access.
	ClubOnly bool `gorm:"not null;default:false"`
	// OrderNum stores the display ordering position.
	OrderNum int `gorm:"not null;default:0"`
	// Images stores page header image identifiers as comma-separated.
	Images string `gorm:"type:text;not null;default:''"`
	// Texts stores page body text blocks as comma-separated.
	Texts string `gorm:"type:text;not null;default:''"`
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// TableName returns the PostgreSQL table name for Page.
func (Page) TableName() string { return "catalog_pages" }
