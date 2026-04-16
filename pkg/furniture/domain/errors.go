package domain

import "errors"

// ErrDefinitionNotFound defines missing item definition lookup behavior.
var ErrDefinitionNotFound = errors.New("item definition not found")

// ErrItemNotFound defines missing item instance lookup behavior.
var ErrItemNotFound = errors.New("item not found")

// ErrItemNotOwned defines unauthorized item ownership behavior.
var ErrItemNotOwned = errors.New("item not owned by user")

// ErrItemNotTradable defines trade restriction behavior.
var ErrItemNotTradable = errors.New("item is not tradable")

// ErrItemNotGiftable defines gift restriction behavior.
var ErrItemNotGiftable = errors.New("item is not giftable")

// ErrItemNotRecyclable defines recycle restriction behavior.
var ErrItemNotRecyclable = errors.New("item is not recyclable")

// ErrItemNotExchangeable defines exchange restriction behavior.
var ErrItemNotExchangeable = errors.New("item is not exchangeable")

// ErrLimitedSoldOut defines limited edition stock exhaustion behavior.
var ErrLimitedSoldOut = errors.New("limited edition sold out")

// ErrInvalidInteraction defines unsupported interaction behavior.
var ErrInvalidInteraction = errors.New("item interaction is not supported")

// ErrItemNotPlaced defines behavior for room-only interactions on inventory items.
var ErrItemNotPlaced = errors.New("item is not placed in a room")
