package http

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestProtectWithAPIKeyEnforcesHeader verifies unauthorized and authorized flows.
func TestProtectWithAPIKeyEnforcesHeader(t *testing.T) {
	module := New(Options{})
	if err := module.ProtectWithAPIKey("secret", ""); err != nil {
		t.Fatalf("expected api key middleware setup success, got %v", err)
	}
	module.RegisterGET("/secure", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	cases := []struct {
		name       string
		header     string
		query      string
		statusCode int
	}{
		{name: "missing", statusCode: nethttp.StatusUnauthorized},
		{name: "wrong", header: "wrong", statusCode: nethttp.StatusUnauthorized},
		{name: "header", header: "secret", statusCode: nethttp.StatusOK},
		{name: "query", query: "api_key=secret", statusCode: nethttp.StatusOK},
	}
	for _, item := range cases {
		path := "/secure"
		if item.query != "" {
			path += "?" + item.query
		}
		request := httptest.NewRequest(nethttp.MethodGet, path, nil)
		if item.header != "" {
			request.Header.Set(DefaultAPIKeyHeader, item.header)
		}
		response, err := module.App().Test(request)
		if err != nil {
			t.Fatalf("expected request success for %s, got %v", item.name, err)
		}
		if response.StatusCode != item.statusCode {
			t.Fatalf("expected status %d for %s, got %d", item.statusCode, item.name, response.StatusCode)
		}
	}
}

// TestProtectWithAPIKeyRejectsEmptyKey verifies middleware precondition checks.
func TestProtectWithAPIKeyRejectsEmptyKey(t *testing.T) {
	module := New(Options{})
	if err := module.ProtectWithAPIKey("", ""); err == nil {
		t.Fatalf("expected api key middleware failure for empty key")
	}
}
