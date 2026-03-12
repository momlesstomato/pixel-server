package httpapi

import "testing"

// TestOpenAPIPathsIncludesSSOEndpoint verifies authentication OpenAPI path mapping.
func TestOpenAPIPathsIncludesSSOEndpoint(t *testing.T) {
	paths := OpenAPIPaths()
	if paths["/api/v1/sso"] == nil {
		t.Fatalf("expected /api/v1/sso path item in OpenAPI paths")
	}
}
