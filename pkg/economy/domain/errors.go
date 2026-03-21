package domain

import "errors"

// ErrOfferNotFound defines missing marketplace offer lookup behavior.
var ErrOfferNotFound = errors.New("marketplace offer not found")

// ErrOfferNotOpen defines non-open marketplace offer behavior.
var ErrOfferNotOpen = errors.New("marketplace offer is not open")

// ErrSelfPurchase defines buyer equals seller behavior.
var ErrSelfPurchase = errors.New("cannot purchase own marketplace offer")

// ErrMarketplaceDisabled defines disabled marketplace behavior.
var ErrMarketplaceDisabled = errors.New("marketplace is disabled")

// ErrPriceBelowMinimum defines below-minimum listing price behavior.
var ErrPriceBelowMinimum = errors.New("listing price below minimum")

// ErrPriceAboveMaximum defines above-maximum listing price behavior.
var ErrPriceAboveMaximum = errors.New("listing price above maximum")

// ErrMaxOffersReached defines per-user offer limit behavior.
var ErrMaxOffersReached = errors.New("maximum active offers reached")

// ErrItemNotMarketable defines non-marketable item behavior.
var ErrItemNotMarketable = errors.New("item is not marketable")

// ErrTradeLogNotFound defines missing trade log lookup behavior.
var ErrTradeLogNotFound = errors.New("trade log not found")
