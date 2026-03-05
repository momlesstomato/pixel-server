package postgres

import (
	"context"
	"testing"
)

// TestNewInvalidURL ensures invalid DSNs are rejected.
func TestNewInvalidURL(t *testing.T) {
	_, err := New(context.Background(), Config{URL: ":", MinConns: 1, MaxConns: 2})
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

// TestNewAndPingFailure validates lazy connection behavior and ping failure.
func TestNewAndPingFailure(t *testing.T) {
	cfg := Config{URL: "postgres://user:pass@127.0.0.1:1/db?sslmode=disable", MinConns: 1, MaxConns: 2}
	svc, err := New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected service creation, got %v", err)
	}
	defer svc.Close()
	if err := svc.Ping(context.Background()); err == nil {
		t.Fatalf("expected ping failure")
	}
	if svc.Pool() == nil {
		t.Fatalf("expected non-nil pool")
	}
}
