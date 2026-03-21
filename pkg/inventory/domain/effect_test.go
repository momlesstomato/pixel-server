package domain

import (
	"testing"
	"time"
)

// TestEffectIsExpiredPermanent verifies permanent effects never expire.
func TestEffectIsExpiredPermanent(t *testing.T) {
	effect := Effect{IsPermanent: true, Duration: 60}
	if effect.IsExpired() {
		t.Fatalf("expected permanent effect to not be expired")
	}
}

// TestEffectIsExpiredInactive verifies inactive effects are not expired.
func TestEffectIsExpiredInactive(t *testing.T) {
	effect := Effect{Duration: 60}
	if effect.IsExpired() {
		t.Fatalf("expected inactive effect to not be expired")
	}
}

// TestEffectIsExpiredElapsed verifies expired effects are detected.
func TestEffectIsExpiredElapsed(t *testing.T) {
	past := time.Now().Add(-120 * time.Second)
	effect := Effect{Duration: 60, ActivatedAt: &past}
	if !effect.IsExpired() {
		t.Fatalf("expected elapsed effect to be expired")
	}
}

// TestEffectIsExpiredStillActive verifies active effects are not expired.
func TestEffectIsExpiredStillActive(t *testing.T) {
	recent := time.Now().Add(-10 * time.Second)
	effect := Effect{Duration: 60, ActivatedAt: &recent}
	if effect.IsExpired() {
		t.Fatalf("expected active effect to not be expired")
	}
}
