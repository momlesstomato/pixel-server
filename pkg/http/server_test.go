package httpserver

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"pixelsv/pkg/core/transport/local"
)

// TestServerRoutes validates core route behavior.
func TestServerRoutes(t *testing.T) {
	cfg := Config{
		Address:               ":0",
		DisableStartupMessage: true,
		ReadTimeoutSeconds:    10,
		OpenAPIPath:           "/openapi.json",
		SwaggerPath:           "/swagger",
		APIKey:                "secret",
	}
	server, err := New(cfg, nil, local.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	tests := []struct {
		name       string
		path       string
		apiKey     string
		statusCode int
		contains   string
	}{
		{name: "health", path: "/health", statusCode: 200, contains: `"status":"ok"`},
		{name: "ready", path: "/ready", statusCode: 200, contains: `"status":"ready"`},
		{name: "openapi", path: "/openapi.json", statusCode: 200, contains: `"openapi":"3.1.0"`},
		{name: "swagger", path: "/swagger", statusCode: 200, contains: `SwaggerUIBundle`},
		{name: "admin missing key", path: "/api/v1/admin/ping", statusCode: 401, contains: `"error":"missing api key"`},
		{name: "admin valid key", path: "/api/v1/admin/ping", apiKey: "secret", statusCode: 200, contains: `"scope":"admin"`},
		{name: "ws without upgrade", path: "/ws", statusCode: 426, contains: ``},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}
			resp, err := server.App().Test(req)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if resp.StatusCode != tt.statusCode {
				t.Fatalf("expected %d, got %d", tt.statusCode, resp.StatusCode)
			}
			body, _ := io.ReadAll(resp.Body)
			if tt.contains != "" && !strings.Contains(string(body), tt.contains) {
				t.Fatalf("expected body to contain %q, got %q", tt.contains, string(body))
			}
		})
	}
}
