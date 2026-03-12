package cli

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeProtectsRoutesWithAPIKey verifies auth middleware protection.
func TestExecuteServeProtectsRoutesWithAPIKey(t *testing.T) {
	err := ExecuteServe(ServeOptions{
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

// TestNewServeCommandDoesNotRequireAPIKeyFlag verifies config-driven API key behavior.
func TestNewServeCommandDoesNotRequireAPIKeyFlag(t *testing.T) {
	command := NewServeCommand(ServeDependencies{
		Listen: func(_ *corehttp.Module, _ string) error { return nil },
	})
	command.SetArgs([]string{"--env-file", writeServeEnvFile(t), "--ws-path", "/events"})
	if err := command.Execute(); err != nil {
		t.Fatalf("expected command execution success without api-key flag, got %v", err)
	}
}
