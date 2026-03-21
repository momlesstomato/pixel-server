package domain

import "errors"

// ErrPageNotFound defines missing catalog page lookup behavior.
var ErrPageNotFound = errors.New("catalog page not found")

// ErrOfferNotFound defines missing catalog offer lookup behavior.
var ErrOfferNotFound = errors.New("catalog offer not found")

// ErrOfferInactive defines inactive catalog offer behavior.
var ErrOfferInactive = errors.New("catalog offer is inactive")

// ErrPageDisabled defines disabled catalog page behavior.
var ErrPageDisabled = errors.New("catalog page is disabled")

// ErrInsufficientRank defines rank restriction behavior.
var ErrInsufficientRank = errors.New("insufficient rank for catalog page")

// ErrClubRequired defines club membership restriction behavior.
var ErrClubRequired = errors.New("club membership required")

// ErrVoucherNotFound defines missing voucher lookup behavior.
var ErrVoucherNotFound = errors.New("voucher not found")

// ErrVoucherExhausted defines exhausted voucher behavior.
var ErrVoucherExhausted = errors.New("voucher has been fully redeemed")

// ErrVoucherAlreadyRedeemed defines per-user duplicate redemption behavior.
var ErrVoucherAlreadyRedeemed = errors.New("voucher already redeemed by user")

// ErrVoucherDisabled defines disabled voucher behavior.
var ErrVoucherDisabled = errors.New("voucher is disabled")

// ErrRecipientNotFound defines missing gift recipient behavior.
var ErrRecipientNotFound = errors.New("gift recipient not found")

// ErrPurchaseCooldown defines purchase rate limit behavior.
var ErrPurchaseCooldown = errors.New("purchase on cooldown")
