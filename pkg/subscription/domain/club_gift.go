package domain

import (
	"context"
	"time"
)

// ItemDeliverer creates one furniture item in a user's inventory.
type ItemDeliverer interface {
	// DeliverItem creates one furniture instance for the given user.
	DeliverItem(ctx context.Context, userID int, defID int, extraData string, limitedNumber int, limitedTotal int) (int, error)
}

// ClubGift defines one redeemable HC club gift option.
type ClubGift struct {
	// ID stores the stable club gift identifier.
	ID int
	// Name stores the client-visible gift name and selection key.
	Name string
	// ItemDefinitionID stores the delivered furniture definition identifier.
	ItemDefinitionID int
	// SpriteID stores the furniture sprite used by the client preview.
	SpriteID int
	// ExtraData stores delivered item custom data.
	ExtraData string
	// DaysRequired stores the required HC age in days to unlock this gift.
	DaysRequired int
	// VIPOnly stores whether this gift is restricted to VIP-level club users.
	VIPOnly bool
	// Enabled stores whether this gift can currently be claimed.
	Enabled bool
	// OrderNum stores display ordering.
	OrderNum int
	// CreatedAt stores row creation timestamp.
	CreatedAt time.Time
	// UpdatedAt stores row update timestamp.
	UpdatedAt time.Time
}

// ClubGiftInfo defines the current HC gift selector state for one user.
type ClubGiftInfo struct {
	// DaysUntilNextGift stores the remaining days until the next gift becomes available.
	DaysUntilNextGift int
	// GiftsAvailable stores the number of currently claimable gifts.
	GiftsAvailable int
	// ActiveDays stores the current HC age in days.
	ActiveDays int
	// Gifts stores all configured club gift options.
	Gifts []ClubGift
}

// ClubGiftClaimResult defines one successful HC gift claim.
type ClubGiftClaimResult struct {
	// Gift stores the claimed gift definition.
	Gift ClubGift
	// ItemID stores the delivered furniture item identifier.
	ItemID int
}