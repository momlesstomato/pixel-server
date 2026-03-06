package httpserver

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestOpenAPISpecJSON validates required parts of the generated spec.
func TestOpenAPISpecJSON(t *testing.T) {
	var payload map[string]any
	body := OpenAPISpecJSON()
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("expected valid json, got %v", err)
	}
	if payload["openapi"] != "3.1.0" {
		t.Fatalf("unexpected openapi version: %v", payload["openapi"])
	}
}

// TestSwaggerHTML validates Swagger UI page rendering.
func TestSwaggerHTML(t *testing.T) {
	html := SwaggerHTML("/openapi.json")
	if !strings.Contains(html, "SwaggerUIBundle") {
		t.Fatalf("expected swagger ui bundle script")
	}
	if !strings.Contains(html, "/openapi.json") {
		t.Fatalf("expected openapi path")
	}
}
