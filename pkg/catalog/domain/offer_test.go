package domain

import "testing"

// TestCatalogOfferIsLimited verifies limited detection behavior.
func TestCatalogOfferIsLimited(t *testing.T) {
	unlimited := CatalogOffer{LimitedTotal: 0}
	if unlimited.IsLimited() {
		t.Fatalf("expected unlimited offer to not be limited")
	}
	limited := CatalogOffer{LimitedTotal: 100}
	if !limited.IsLimited() {
		t.Fatalf("expected limited offer to be limited")
	}
}

// TestCatalogOfferHasStock verifies stock availability behavior.
func TestCatalogOfferHasStock(t *testing.T) {
	unlimited := CatalogOffer{LimitedTotal: 0}
	if !unlimited.HasStock() {
		t.Fatalf("expected unlimited offer to have stock")
	}
	available := CatalogOffer{LimitedTotal: 100, LimitedSells: 50}
	if !available.HasStock() {
		t.Fatalf("expected partially sold offer to have stock")
	}
	soldOut := CatalogOffer{LimitedTotal: 100, LimitedSells: 100}
	if soldOut.HasStock() {
		t.Fatalf("expected sold out offer to not have stock")
	}
}
