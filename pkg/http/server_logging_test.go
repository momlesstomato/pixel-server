package httpserver

import (
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"pixelsv/pkg/core/transport/local"
)

// TestRequestLogsDisabledOutsideDebug validates that request logs are skipped above debug level.
func TestRequestLogsDisabledOutsideDebug(t *testing.T) {
	core, observed := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	cfg := Config{Address: ":0", DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}
	srv, err := New(cfg, logger, local.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	req := httptest.NewRequest("GET", "/health", nil)
	if _, err := srv.App().Test(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if observed.Len() != 0 {
		t.Fatalf("expected zero request logs at info level, got %d", observed.Len())
	}
}

// TestRequestLogsEnabledAtDebug validates that request logs are emitted at debug level.
func TestRequestLogsEnabledAtDebug(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	cfg := Config{Address: ":0", DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}
	srv, err := New(cfg, logger, local.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	req := httptest.NewRequest("GET", "/health", nil)
	if _, err := srv.App().Test(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if observed.Len() == 0 {
		t.Fatalf("expected request logs at debug level")
	}
}

// Test404ClientErrorIsNotErrorLevel validates 404 logs do not use error severity.
func Test404ClientErrorIsNotErrorLevel(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)
	cfg := Config{Address: ":0", DisableStartupMessage: true, ReadTimeoutSeconds: 10, OpenAPIPath: "/openapi.json", SwaggerPath: "/swagger", APIKey: "secret"}
	srv, err := New(cfg, logger, local.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	if _, err := srv.App().Test(req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	for _, entry := range observed.All() {
		if entry.Level == zapcore.ErrorLevel {
			t.Fatalf("expected no error-level logs for 404, got %s", entry.Message)
		}
	}
}
