package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	coreapp "github.com/momlesstomato/pixel-server/core/app"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/authentication"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TestRegisterRoutesIssuesTicket verifies route registration and issuance behavior.
func TestRegisterRoutesIssuesTicket(t *testing.T) {
	server := startMiniRedis(t)
	client := redislib.NewClient(&redislib.Options{Addr: server.Addr()})
	store, err := authentication.NewRedisStore(client, "sso")
	if err != nil {
		t.Fatalf("expected store creation success, got %v", err)
	}
	service := authentication.NewService(store, authentication.Config{DefaultTTLSeconds: 300, MaxTTLSeconds: 1800, KeyPrefix: "sso"})
	module, err := corehttp.Initializer{}.InitializeHTTP(coreapp.Config{APIKey: "secret"}, zap.NewNop())
	if err != nil {
		t.Fatalf("expected http initializer success, got %v", err)
	}
	if err := RegisterRoutes(module, service); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	body := bytes.NewBufferString(`{"user_id":12,"ttl_seconds":120}`)
	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/sso", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(corehttp.DefaultAPIKeyHeader, "secret")
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	var payload map[string]string
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("expected decode success, got %v", err)
	}
	if payload["ticket"] == "" || payload["expires_at"] == "" {
		t.Fatalf("unexpected response payload: %+v", payload)
	}
	if value, getErr := server.Get("sso:" + payload["ticket"]); getErr != nil || value != "12" {
		t.Fatalf("expected redis ticket persistence, got value=%q err=%v", value, getErr)
	}
}

// TestRegisterRoutesRejectsInvalidPayload verifies request validation behavior.
func TestRegisterRoutesRejectsInvalidPayload(t *testing.T) {
	module, err := corehttp.Initializer{}.InitializeHTTP(coreapp.Config{APIKey: "secret"}, zap.NewNop())
	if err != nil {
		t.Fatalf("expected http initializer success, got %v", err)
	}
	service := &stubIssuer{}
	if err := RegisterRoutes(module, service); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/sso", bytes.NewBufferString(`{"user_id":0}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(corehttp.DefaultAPIKeyHeader, "secret")
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

// TestRegisterRoutesRejectsNilDependencies verifies registration precondition checks.
func TestRegisterRoutesRejectsNilDependencies(t *testing.T) {
	if err := RegisterRoutes(nil, &stubIssuer{}); err == nil {
		t.Fatalf("expected registration failure for nil module")
	}
	module := corehttp.New(corehttp.Options{})
	if err := RegisterRoutes(module, nil); err == nil {
		t.Fatalf("expected registration failure for nil issuer")
	}
}

// stubIssuer defines deterministic issue behavior for route tests.
type stubIssuer struct{}

// Issue issues one deterministic ticket payload.
func (issuer *stubIssuer) Issue(_ context.Context, request authentication.IssueRequest) (authentication.IssueResult, error) {
	if request.UserID <= 0 {
		return authentication.IssueResult{}, fmt.Errorf("user id must be positive")
	}
	return authentication.IssueResult{Ticket: "t", ExpiresAt: time.Unix(1, 0)}, nil
}

// startMiniRedis creates an isolated miniredis instance for tests.
func startMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup, got %v", err)
	}
	t.Cleanup(server.Close)
	return server
}
