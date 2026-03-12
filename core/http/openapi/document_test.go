package openapi

import (
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestBuildDocumentIncludesPaths verifies document path composition behavior.
func TestBuildDocumentIncludesPaths(t *testing.T) {
	document := BuildDocument("/socket", map[string]any{
		"/api/v1/sso": map[string]any{"post": map[string]any{"summary": "issue"}},
	})
	paths := document["paths"].(map[string]any)
	if paths["/socket"] == nil || paths["/api/v1/sso"] == nil {
		t.Fatalf("expected websocket and extra paths in document: %+v", paths)
	}
	specOperation := paths[DefaultSpecPath].(map[string]any)["get"].(map[string]any)
	if len(specOperation["security"].([]any)) != 0 {
		t.Fatalf("expected openapi json route to be public")
	}
}

// TestRegisterRoutesServesSpecAndUI verifies route registration and responses.
func TestRegisterRoutesServesSpecAndUI(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	document := BuildDocument("/ws", map[string]any{})
	if err := RegisterRoutes(module, document, "", ""); err != nil {
		t.Fatalf("expected route registration success, got %v", err)
	}
	specRequest := httptest.NewRequest(nethttp.MethodGet, DefaultSpecPath, nil)
	specResponse, err := module.App().Test(specRequest)
	if err != nil {
		t.Fatalf("expected spec request success, got %v", err)
	}
	if specResponse.StatusCode != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", specResponse.StatusCode)
	}
	uiRequest := httptest.NewRequest(nethttp.MethodGet, DefaultUIPath, nil)
	uiResponse, err := module.App().Test(uiRequest)
	if err != nil {
		t.Fatalf("expected ui request success, got %v", err)
	}
	if uiResponse.StatusCode != nethttp.StatusOK {
		t.Fatalf("expected status 200, got %d", uiResponse.StatusCode)
	}
}

// TestRegisterRoutesRejectsNilModule verifies registration precondition checks.
func TestRegisterRoutesRejectsNilModule(t *testing.T) {
	if err := RegisterRoutes(nil, map[string]any{}, "", ""); err == nil {
		t.Fatalf("expected registration failure for nil module")
	}
}

// TestSwaggerHTMLIncludesSpecPath verifies ui rendering behavior.
func TestSwaggerHTMLIncludesSpecPath(t *testing.T) {
	page := swaggerHTML("/openapi.json")
	if !strings.Contains(page, "/openapi.json") {
		t.Fatalf("expected html to reference spec path, got %s", page)
	}
}
