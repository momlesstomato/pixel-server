package domain

// CatalogOffer defines one purchasable offer within a catalog page.
type CatalogOffer struct {
	// ID stores stable offer identifier.
	ID int
	// PageID stores the owning page identifier.
	PageID int
	// ItemDefinitionID stores the furniture definition foreign key.
	ItemDefinitionID int
	// CatalogName stores the client-visible offer name.
	CatalogName string
	// CostPrimary stores the primary currency price component.
	CostPrimary int
	// CostPrimaryType stores the primary currency type identifier.
	CostPrimaryType int
	// CostSecondary stores the secondary currency price component.
	CostSecondary int
	// CostSecondaryType stores the secondary currency type identifier.
	CostSecondaryType int
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
