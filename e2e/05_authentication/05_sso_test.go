package authentication

import (
	"bytes"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/momlesstomato/pixel-server/core/cli"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
)

// issueResponse defines SSO issue response payload.
type issueResponse struct {
	// Ticket stores issued SSO token.
	Ticket string `json:"ticket"`
}

// Test05SSOIssuesTicketWithAPIKey verifies SSO issuance behavior.
func Test05SSOIssuesTicketWithAPIKey(t *testing.T) {
	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("expected miniredis startup success, got %v", err)
	}
	defer redisServer.Close()
	envFile := testkit.WriteServeEnvFile(t, redisServer.Addr())
	err = cli.ExecuteServe(cli.ServeOptions{EnvFile: envFile}, func(module *corehttp.Module, _ string) error {
		body := []byte(`{"user_id":7,"ttl_seconds":60}`)
		request := httptest.NewRequest(nethttp.MethodPost, "/api/v1/sso", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(corehttp.DefaultAPIKeyHeader, "test-key")
		response, testErr := module.App().Test(request)
		if testErr != nil {
			return testErr
		}
		if response.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected SSO issue status 200, got %d", response.StatusCode)
		}
		var payload issueResponse
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			t.Fatalf("expected json decode success, got %v", err)
		}
		if payload.Ticket == "" {
			t.Fatalf("expected non-empty issued ticket")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve SSO flow success, got %v", err)
	}
}
