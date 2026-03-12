package tests

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestProtectWithAPIKeyEnforcesHeader verifies unauthorized and authorized flows.
func TestProtectWithAPIKeyEnforcesHeader(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
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
			request.Header.Set(corehttp.DefaultAPIKeyHeader, item.header)
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
	module := corehttp.New(corehttp.Options{})
	if err := module.ProtectWithAPIKey("", ""); err == nil {
		t.Fatalf("expected api key middleware failure for empty key")
	}
}

// TestProtectWithAPIKeyAllowsOpenAPIDocs verifies docs route bypass behavior.
func TestProtectWithAPIKeyAllowsOpenAPIDocs(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.ProtectWithAPIKey("secret", ""); err != nil {
		t.Fatalf("expected api key middleware setup success, got %v", err)
	}
	module.RegisterGET(corehttp.DefaultOpenAPISpecPath, func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	module.RegisterGET(corehttp.DefaultSwaggerUIPath, func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})
	specRequest := httptest.NewRequest(nethttp.MethodGet, corehttp.DefaultOpenAPISpecPath, nil)
	specResponse, specErr := module.App().Test(specRequest)
	if specErr != nil {
		t.Fatalf("expected spec request success, got %v", specErr)
	}
	if specResponse.StatusCode != nethttp.StatusOK {
		t.Fatalf("expected status 200 for spec, got %d", specResponse.StatusCode)
	}
	uiRequest := httptest.NewRequest(nethttp.MethodGet, corehttp.DefaultSwaggerUIPath, nil)
	uiResponse, uiErr := module.App().Test(uiRequest)
	if uiErr != nil {
		t.Fatalf("expected ui request success, got %v", uiErr)
	}
	if uiResponse.StatusCode != nethttp.StatusOK {
		t.Fatalf("expected status 200 for ui, got %d", uiResponse.StatusCode)
	}
}

// TestProtectWithAPIKeyAllowsWebSocketRoute verifies websocket route bypass behavior.
func TestProtectWithAPIKeyAllowsWebSocketRoute(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.ProtectWithAPIKey("secret", ""); err != nil {
		t.Fatalf("expected api key middleware setup success, got %v", err)
	}
	if err := module.RegisterWebSocket("/ws", func(_ *websocket.Conn) {}); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	request := httptest.NewRequest(nethttp.MethodGet, "/ws", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusUpgradeRequired {
		t.Fatalf("expected status 426 for websocket path, got %d", response.StatusCode)
	}
}

// TestProtectWithAPIKeyAllowsWebSocketTrailingSlash verifies normalized websocket bypass behavior.
func TestProtectWithAPIKeyAllowsWebSocketTrailingSlash(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	if err := module.ProtectWithAPIKey("secret", ""); err != nil {
		t.Fatalf("expected api key middleware setup success, got %v", err)
	}
	if err := module.RegisterWebSocket("/ws", func(_ *websocket.Conn) {}); err != nil {
		t.Fatalf("expected websocket registration success, got %v", err)
	}
	request := httptest.NewRequest(nethttp.MethodGet, "/ws/", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusUpgradeRequired {
		t.Fatalf("expected status 426 for websocket path, got %d", response.StatusCode)
	}
}
