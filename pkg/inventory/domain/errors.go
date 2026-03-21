package domain

import "errors"

// ErrBadgeNotFound defines missing badge lookup behavior.
var ErrBadgeNotFound = errors.New("badge not found")

// ErrBadgeAlreadyOwned defines duplicate badge award behavior.
var ErrBadgeAlreadyOwned = errors.New("badge already owned")

// ErrBadgeSlotInvalid defines out-of-range badge slot behavior.
var ErrBadgeSlotInvalid = errors.New("badge slot is invalid")

// ErrEffectNotFound defines missing effect lookup behavior.
var ErrEffectNotFound = errors.New("effect not found")

// ErrEffectAlreadyOwned defines duplicate effect award behavior.
var ErrEffectAlreadyOwned = errors.New("effect already owned")

// ErrInsufficientCredits defines credit underflow behavior.
var ErrInsufficientCredits = errors.New("insufficient credits")

// ErrInsufficientCurrency defines activity-point underflow behavior.
var ErrInsufficientCurrency = errors.New("insufficient currency")

// ErrCurrencyTypeUnknown defines unregistered currency type behavior.
var ErrCurrencyTypeUnknown = errors.New("unknown currency type")

// ErrInventoryFull defines inventory capacity overflow behavior.
var ErrInventoryFull = errors.New("inventory is full")
