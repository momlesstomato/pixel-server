package redis

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

// TestServiceNewAndPing validates client creation and ping.
func TestServiceNewAndPing(t *testing.T) {
	server := miniredis.RunT(t)
	cfg := Config{URL: "redis://" + server.Addr(), KeyPrefix: "pixelsv", SessionTTLSeconds: 30}
	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer svc.Close()
	if err := svc.Ping(context.Background()); err != nil {
		t.Fatalf("expected ping success, got %v", err)
	}
}

// TestServiceNewInvalidURL validates invalid URL rejection.
func TestServiceNewInvalidURL(t *testing.T) {
	_, err := New(Config{URL: ":", KeyPrefix: "px", SessionTTLSeconds: 10})
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

// TestDebugLoggingEnabled validates LOG_LEVEL debug detection.
func TestDebugLoggingEnabled(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	if !debugLoggingEnabled() {
		t.Fatalf("expected debug logging enabled")
	}
	t.Setenv("LOG_LEVEL", "warn")
	if debugLoggingEnabled() {
		t.Fatalf("expected debug logging disabled")
	}
}
