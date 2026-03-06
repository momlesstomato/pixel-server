package httpserver

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestAPIKeyMiddleware validates missing, invalid, and valid key behavior.
func TestAPIKeyMiddleware(t *testing.T) {
	app := fiber.New()
	app.Get("/admin", APIKeyMiddleware("secret"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	reqMissing := httptest.NewRequest("GET", "/admin", nil)
	respMissing, _ := app.Test(reqMissing)
	if respMissing.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", respMissing.StatusCode)
	}
	reqInvalid := httptest.NewRequest("GET", "/admin", nil)
	reqInvalid.Header.Set("X-API-Key", "nope")
	respInvalid, _ := app.Test(reqInvalid)
	if respInvalid.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", respInvalid.StatusCode)
	}
	reqValid := httptest.NewRequest("GET", "/admin", nil)
	reqValid.Header.Set("X-API-Key", "secret")
	respValid, _ := app.Test(reqValid)
	if respValid.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", respValid.StatusCode)
	}
}
