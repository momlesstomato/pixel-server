package startup

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/momlesstomato/pixel-server/core/cli"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
)

// Test01StartupServeBootstrapsDocsAndWebSocket verifies startup flow for docs and websocket upgrade routes.
func Test01StartupServeBootstrapsDocsAndWebSocket(t *testing.T) {
	err := cli.ExecuteServe(cli.ServeOptions{EnvFile: testkit.WriteServeEnvFile(t, ""), WebSocketPath: "/ws"}, func(module *corehttp.Module, _ string) error {
		specRequest := httptest.NewRequest(nethttp.MethodGet, "/openapi.json", nil)
		specResponse, specErr := module.App().Test(specRequest)
		if specErr != nil {
			return specErr
		}
		if specResponse.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected openapi status 200, got %d", specResponse.StatusCode)
		}
		webSocketRequest := httptest.NewRequest(nethttp.MethodGet, "/ws", nil)
		webSocketResponse, wsErr := module.App().Test(webSocketRequest)
		if wsErr != nil {
			return wsErr
		}
		if webSocketResponse.StatusCode != nethttp.StatusUpgradeRequired {
			t.Fatalf("expected websocket status 426, got %d", webSocketResponse.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve startup success, got %v", err)
	}
}
