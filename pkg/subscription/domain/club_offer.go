package domain

// ClubOffer defines one purchasable club membership option.
type ClubOffer struct {
	// ID stores stable club offer identifier.
	ID int
	// Name stores the offer display name.
	Name string
	// Days stores the membership duration in days.
	Days int
	// Credits stores the credit price.
	Credits int
	// Points stores the activity-point price.
	Points int
	// PointsType stores the activity-point currency type.
	PointsType int
	// OfferType stores the membership tier key.
	OfferType string
	// Giftable stores whether the offer can be gifted.
	Giftable bool
	// Enabled stores whether the offer is currently purchasable.
	Enabled bool
}
