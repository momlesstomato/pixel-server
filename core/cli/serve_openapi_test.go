package cli

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeRegistersOpenAPIRoutes verifies docs endpoint registration behavior.
func TestExecuteServeRegistersOpenAPIRoutes(t *testing.T) {
	err := ExecuteServe(ServeOptions{
		EnvFile: writeServeEnvFile(t), WebSocketPath: "/realtime",
	}, func(module *corehttp.Module, _ string) error {
		specRequest := httptest.NewRequest(nethttp.MethodGet, "/openapi.json", nil)
		specResponse, specErr := module.App().Test(specRequest)
		if specErr != nil {
			return specErr
		}
		if specResponse.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected status 200, got %d", specResponse.StatusCode)
		}
		uiRequest := httptest.NewRequest(nethttp.MethodGet, "/swagger", nil)
		uiResponse, uiErr := module.App().Test(uiRequest)
		if uiErr != nil {
			return uiErr
		}
		if uiResponse.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected status 200, got %d", uiResponse.StatusCode)
		}
		apiRequest := httptest.NewRequest(nethttp.MethodPost, "/api/v1/sso", nil)
		apiResponse, apiErr := module.App().Test(apiRequest)
		if apiErr != nil {
			return apiErr
		}
		if apiResponse.StatusCode != nethttp.StatusUnauthorized {
			t.Fatalf("expected status 401 for protected endpoint, got %d", apiResponse.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve execution success, got %v", err)
	}
}
