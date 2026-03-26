package domain

import "context"

// ActivityCurrencyValidator is a secondary port used to validate that an
// activity-point type identifier is registered and enabled in the currency
// registry before persisting a catalog offer.
type ActivityCurrencyValidator interface {
	// IsValidActivityPointType reports whether typeID is registered and enabled.
	IsValidActivityPointType(ctx context.Context, typeID int) (bool, error)
}

// CatalogOffer defines one purchasable offer within a catalog page.
type CatalogOffer struct {
	// ID stores stable offer identifier.
	ID int
	// PageID stores the owning page identifier.
	PageID int
	// ItemDefinitionID stores the furniture definition foreign key.
	ItemDefinitionID int
	// SpriteID stores the furniture sprite identifier resolved from the
	// linked item definition at read time. Used by the client to render
	// the item icon and preview in the catalog.
	SpriteID int
	// ItemType stores the furniture type code resolved from the linked
	// item definition at read time ("s" floor, "i" wall, "e" effect).
	ItemType string
	// CatalogName stores the client-visible offer display name.
	// This field is never stored in the database; it is always resolved
	// at read time from the linked item definition's public_name.
	CatalogName string
	// CostCredits stores the credits price component (can be zero for activity-point-only offers).
	CostCredits int
	// CostActivityPoints stores the activity-point price component (can be zero).
	CostActivityPoints int
	// ActivityPointType stores the activity-point currency type identifier.
	ActivityPointType int
	// Amount stores number of items per single purchase.
	Amount int
	// LimitedTotal stores total limited edition print run, zero for unlimited.
	LimitedTotal int
	// LimitedSells stores current limited edition sold count.
	LimitedSells int
	// OfferActive stores whether the offer is currently purchasable.
	OfferActive bool
	// ExtraData stores item-specific custom data payload.
	ExtraData string
	// BadgeID stores optional badge code awarded with purchase.
	BadgeID string
	// ClubOnly stores whether club membership is required.
	ClubOnly bool
	// OrderNum stores display sort position.
	OrderNum int
}

// IsLimited reports whether this offer is a limited edition.
func (o CatalogOffer) IsLimited() bool {
	return o.LimitedTotal > 0
}

// HasStock reports whether limited stock remains.
func (o CatalogOffer) HasStock() bool {
	if !o.IsLimited() {
		return true
	}
	return o.LimitedSells < o.LimitedTotal
}
