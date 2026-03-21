package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// VoucherRedeemed fires after a voucher is redeemed.
type VoucherRedeemed struct {
	sdk.BaseEvent
	// UserID stores the user identifier.
	UserID int
	// VoucherCode stores the voucher code.
	VoucherCode string
}
