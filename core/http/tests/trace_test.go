package tests

import (
	"bytes"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	corelogging "github.com/momlesstomato/pixel-server/core/logging"
)

// TestTraceMiddlewareSetsRayIDHeader verifies X-Ray-ID header is present on every response.
func TestTraceMiddlewareSetsRayIDHeader(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	module.RegisterGET("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(nethttp.StatusOK)
	})
	request := httptest.NewRequest(nethttp.MethodGet, "/ping", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	rayID := response.Header.Get(corehttp.HeaderRayID)
	if rayID == "" {
		t.Fatalf("expected %s header to be set on response", corehttp.HeaderRayID)
	}
}

// TestTraceMiddlewareGeneratesUniqueIDs verifies each request receives a distinct ray_id.
func TestTraceMiddlewareGeneratesUniqueIDs(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	module.RegisterGET("/ping", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(nethttp.StatusOK)
	})
	first, err := module.App().Test(httptest.NewRequest(nethttp.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatalf("expected first request success, got %v", err)
	}
	second, err := module.App().Test(httptest.NewRequest(nethttp.MethodGet, "/ping", nil))
	if err != nil {
		t.Fatalf("expected second request success, got %v", err)
	}
	if first.Header.Get(corehttp.HeaderRayID) == second.Header.Get(corehttp.HeaderRayID) {
		t.Fatal("expected different ray_id values for distinct requests")
	}
}

// TestErrorHandlerReturnsJSONError verifies error responses use JSON error format.
func TestErrorHandlerReturnsJSONError(t *testing.T) {
	module := corehttp.New(corehttp.Options{})
	module.RegisterGET("/fail", func(_ *fiber.Ctx) error {
		return fiber.NewError(nethttp.StatusBadRequest, "invalid input")
	})
	request := httptest.NewRequest(nethttp.MethodGet, "/fail", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
	var body map[string]any
	if jsonErr := json.NewDecoder(response.Body).Decode(&body); jsonErr != nil {
		t.Fatalf("expected JSON response body, got decode error: %v", jsonErr)
	}
	if body["error"] == "" || body["error"] == nil {
		t.Fatalf("expected error field in response, got %v", body)
	}
	if response.Header.Get(corehttp.HeaderRayID) == "" {
		t.Fatalf("expected %s header on error response", corehttp.HeaderRayID)
	}
}

// TestErrorHandlerLogs5xxWithRayID verifies 5xx errors are logged with ray_id field.
func TestErrorHandlerLogs5xxWithRayID(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := corelogging.New(corelogging.Config{Format: "json", Level: "error"}, buffer)
	if err != nil {
		t.Fatalf("expected logger creation to succeed, got %v", err)
	}
	module := corehttp.New(corehttp.Options{Logger: logger})
	module.RegisterGET("/crash", func(_ *fiber.Ctx) error {
		return fiber.NewError(nethttp.StatusInternalServerError, "boom")
	})
	request := httptest.NewRequest(nethttp.MethodGet, "/crash", nil)
	response, err := module.App().Test(request)
	if err != nil {
		t.Fatalf("expected request success, got %v", err)
	}
	if response.StatusCode != nethttp.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", response.StatusCode)
	}
	output := buffer.String()
	if output == "" {
		t.Fatal("expected log output for 5xx response")
	}
	if !strings.Contains(output, "ray_id") {
		t.Fatalf("expected ray_id field in error log, got: %s", output)
	}
}
