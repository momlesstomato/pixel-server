package cli

import (
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// TestExecuteServeProtectsRoutesWithAPIKey verifies auth middleware protection.
func TestExecuteServeProtectsRoutesWithAPIKey(t *testing.T) {
	err := ExecuteServe(ServeOptions{
		EnvFile: writeServeEnvFile(t), WebSocketPath: "/realtime", APIKey: "test-key",
	}, func(module *corehttp.Module, _ string) error {
		request := httptest.NewRequest(nethttp.MethodGet, "/realtime", nil)
		response, testErr := module.App().Test(request)
		if testErr != nil {
			return testErr
		}
		if response.StatusCode != nethttp.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", response.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve execution success, got %v", err)
	}
}

// TestNewServeCommandRequiresAPIKey verifies required flag validation.
func TestNewServeCommandRequiresAPIKey(t *testing.T) {
	command := NewServeCommand(ServeDependencies{})
	command.SetArgs([]string{"--env-file", writeServeEnvFile(t), "--ws-path", "/events"})
	err := command.Execute()
	if err == nil {
		t.Fatalf("expected command execution failure when api key is missing")
	}
	if !strings.Contains(err.Error(), "api-key") {
		t.Fatalf("expected missing api-key error, got %v", err)
	}
}
