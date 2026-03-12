package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// DefaultSpecPath defines the default OpenAPI JSON route.
const DefaultSpecPath = corehttp.DefaultOpenAPISpecPath

// DefaultUIPath defines the default Swagger UI route.
const DefaultUIPath = corehttp.DefaultSwaggerUIPath

// BuildDocument creates an OpenAPI 3.1 document for configured endpoints.
func BuildDocument(webSocketPath string, extraPaths map[string]any) map[string]any {
	path := strings.TrimSpace(webSocketPath)
	if path == "" {
		path = "/ws"
	}
	paths := map[string]any{
		path: map[string]any{
			"get": map[string]any{
				"tags":        []string{"realtime"},
				"summary":     "WebSocket endpoint",
				"description": "Upgrades to websocket transport for realtime communication.",
				"responses": map[string]any{
					"101": map[string]any{"description": "Switching Protocols"},
					"401": map[string]any{"description": "Invalid API key"},
					"426": map[string]any{"description": "WebSocket upgrade required"},
				},
				"security": []map[string]any{{"ApiKeyAuth": []string{}}},
			},
		},
		DefaultSpecPath: map[string]any{
			"get": map[string]any{
				"tags":        []string{"documentation"},
				"summary":     "OpenAPI document",
				"description": "Returns OpenAPI 3.1 specification for all API endpoints.",
				"responses": map[string]any{
					"200": map[string]any{"description": "OpenAPI specification"},
				},
				"security": []any{},
			},
		},
		DefaultUIPath: map[string]any{
			"get": map[string]any{
				"tags":        []string{"documentation"},
				"summary":     "Swagger UI",
				"description": "Returns Swagger UI page for API exploration.",
				"responses": map[string]any{
					"200": map[string]any{"description": "Swagger UI HTML"},
				},
				"security": []any{},
			},
		},
	}
	for pathKey, pathItem := range extraPaths {
		paths[pathKey] = pathItem
	}
	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   "Pixel Server API",
			"version": "1.0.0",
		},
		"paths": paths,
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"ApiKeyAuth": map[string]any{
					"type": "apiKey", "in": "header", "name": corehttp.DefaultAPIKeyHeader,
				},
			},
		},
		"security": []map[string]any{{"ApiKeyAuth": []string{}}},
	}
}

// RegisterRoutes registers OpenAPI spec and Swagger UI routes.
func RegisterRoutes(module *corehttp.Module, document map[string]any, specPath string, uiPath string) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	route := strings.TrimSpace(specPath)
	if route == "" {
		route = DefaultSpecPath
	}
	uiRoute := strings.TrimSpace(uiPath)
	if uiRoute == "" {
		uiRoute = DefaultUIPath
	}
	module.RegisterGET(route, func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return ctx.Send(encoded)
	})
	module.RegisterGET(uiRoute, func(ctx *fiber.Ctx) error {
		return ctx.Type("html").SendString(swaggerHTML(route))
	})
	return nil
}

// swaggerHTML returns a minimal Swagger UI page bound to one OpenAPI route.
func swaggerHTML(specPath string) string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"><title>Swagger UI</title>" +
		"<link rel=\"stylesheet\" href=\"https://unpkg.com/swagger-ui-dist/swagger-ui.css\"></head>" +
		"<body><div id=\"swagger-ui\"></div><script src=\"https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js\"></script>" +
		"<script>window.ui = SwaggerUIBundle({url: '" + specPath + "',dom_id: '#swagger-ui'});</script></body></html>"
}
