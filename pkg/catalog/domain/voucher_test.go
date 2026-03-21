package domain

import "testing"

// TestVoucherIsExhausted verifies voucher exhaustion detection.
func TestVoucherIsExhausted(t *testing.T) {
	available := Voucher{MaxUses: 10, CurrentUses: 5}
	if available.IsExhausted() {
		t.Fatalf("expected non-exhausted voucher")
	}
	exhausted := Voucher{MaxUses: 10, CurrentUses: 10}
	if !exhausted.IsExhausted() {
		t.Fatalf("expected exhausted voucher")
	}
	over := Voucher{MaxUses: 10, CurrentUses: 15}
	if !over.IsExhausted() {
		t.Fatalf("expected over-exhausted voucher")
	}
}
