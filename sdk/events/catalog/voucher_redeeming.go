package catalog

import sdk "github.com/momlesstomato/pixel-sdk"

// VoucherRedeeming fires before a voucher redemption is committed.
type VoucherRedeeming struct {
	sdk.BaseCancellable
	// UserID stores the redeeming user identifier.
	UserID int
	// VoucherCode stores the voucher code.
	VoucherCode string
	// VoucherID stores the voucher identifier.
	VoucherID int
}
