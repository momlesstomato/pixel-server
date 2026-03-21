package domain

import "errors"

// ErrSubscriptionNotFound defines missing subscription lookup behavior.
var ErrSubscriptionNotFound = errors.New("subscription not found")

// ErrSubscriptionAlreadyActive defines duplicate subscription behavior.
var ErrSubscriptionAlreadyActive = errors.New("subscription already active")

// ErrClubOfferNotFound defines missing club offer lookup behavior.
var ErrClubOfferNotFound = errors.New("club offer not found")

// ErrClubOfferDisabled defines disabled club offer behavior.
var ErrClubOfferDisabled = errors.New("club offer is disabled")

// ErrTargetedOfferNotFound defines missing targeted offer lookup behavior.
var ErrTargetedOfferNotFound = errors.New("targeted offer not found")

// ErrTargetedOfferExpired defines expired targeted offer behavior.
var ErrTargetedOfferExpired = errors.New("targeted offer has expired")

// ErrPurchaseLimitReached defines per-user purchase limit behavior.
var ErrPurchaseLimitReached = errors.New("purchase limit reached")
