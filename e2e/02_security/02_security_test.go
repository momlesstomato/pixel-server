package security

import (
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/momlesstomato/pixel-server/core/cli"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/e2e/testkit"
)

// Test02SecurityProtectsAPIAndLeavesDocsPublic verifies API key protection and docs bypass behavior.
func Test02SecurityProtectsAPIAndLeavesDocsPublic(t *testing.T) {
	err := cli.ExecuteServe(cli.ServeOptions{EnvFile: testkit.WriteServeEnvFile(t, "")}, func(module *corehttp.Module, _ string) error {
		docsRequest := httptest.NewRequest(nethttp.MethodGet, "/swagger", nil)
		docsResponse, docsErr := module.App().Test(docsRequest)
		if docsErr != nil {
			return docsErr
		}
		if docsResponse.StatusCode != nethttp.StatusOK {
			t.Fatalf("expected docs status 200, got %d", docsResponse.StatusCode)
		}
		protectedRequest := httptest.NewRequest(nethttp.MethodPost, "/api/v1/sso", nil)
		protectedResponse, protectedErr := module.App().Test(protectedRequest)
		if protectedErr != nil {
			return protectedErr
		}
		if protectedResponse.StatusCode != nethttp.StatusUnauthorized {
			t.Fatalf("expected protected status 401, got %d", protectedResponse.StatusCode)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected serve security checks success, got %v", err)
	}
}
