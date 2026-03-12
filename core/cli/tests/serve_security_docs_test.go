package tests

import (
	nethttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/momlesstomato/pixel-server/core/cli"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeProtectsRoutesWithAPIKey verifies auth middleware protection.
func TestExecuteServeProtectsRoutesWithAPIKey(t *testing.T) {
	err := cli.ExecuteServe(cli.ServeOptions{
		EnvFile: writeServeEnvFile(t), WebSocketPath: "/realtime",
	}, func(module *corehttp.Module, _ string) error {
		request := httptest.NewRequest(nethttp.MethodGet, "/realtime", nil)
		response, testErr := module.App().Test(request)
		if testErr != nil {
			return testErr
		}
		if response.StatusCode != nethttp.StatusUpgradeRequired {
			t.Fatalf("expected status 426, got %d", response.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve execution success, got %v", err)
	}
}

// TestExecuteServeRegistersOpenAPIRoutes verifies docs endpoint registration behavior.
func TestExecuteServeRegistersOpenAPIRoutes(t *testing.T) {
	err := cli.ExecuteServe(cli.ServeOptions{
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

// TestNewServeCommandDoesNotRequireAPIKeyFlag verifies config-driven API key behavior.
func TestNewServeCommandDoesNotRequireAPIKeyFlag(t *testing.T) {
	command := cli.NewServeCommand(cli.ServeDependencies{
		Listen: func(_ *corehttp.Module, _ string) error { return nil },
	})
	command.SetArgs([]string{"--env-file", writeServeEnvFile(t), "--ws-path", "/events"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected command execution success without api-key flag, got %v", err)
	}
}

// writeServeEnvFile writes a valid env file for serve startup tests.
func writeServeEnvFile(t *testing.T) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), ".env")
	content := []byte("APP_BIND_IP=127.0.0.1\nAPP_PORT=3987\nAPP_API_KEY=test-key\nREDIS_ADDRESS=localhost:6379\nPOSTGRES_DSN=postgres://postgres:postgres@127.0.0.1:5432/pixel_server?sslmode=disable\nUSERS_JWT_SECRET=secret\nLOGGING_LEVEL=debug\n")
	if err := os.WriteFile(filePath, content, 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}
	return filePath
}
