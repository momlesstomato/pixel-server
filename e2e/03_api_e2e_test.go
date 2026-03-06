package e2e_test

import (
	"net/http/httptest"
	"testing"

	"pixelsv/pkg/config"
	httpserver "pixelsv/pkg/http"
	logpkg "pixelsv/pkg/log"
)

// Test03CoreAPIComposition validates config-to-runtime API composition.
func Test03CoreAPIComposition(t *testing.T) {
	t.Setenv("API_KEY", "secret")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "info")
	v, err := config.NewViper(config.DefaultLoadOptions())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := logpkg.BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := httpserver.BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	logCfg, err := logpkg.FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	httpCfg, err := httpserver.FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	logger, err := logpkg.New(logCfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer logger.Sync()
	srv, err := httpserver.New(httpCfg, logger)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	req := httptest.NewRequest("GET", "/api/v1/admin/ping", nil)
	req.Header.Set("X-API-Key", "secret")
	resp, err := srv.App().Test(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
