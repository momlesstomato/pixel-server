package httpserver

import "fmt"

// OpenAPISpecJSON returns the OpenAPI JSON specification payload.
func OpenAPISpecJSON() []byte {
	return []byte(`{"openapi":"3.1.0","info":{"title":"pixelsv Core API","version":"0.1.0"},"servers":[{"url":"/"}],"paths":{"/health":{"get":{"summary":"Liveness probe","tags":["core"],"responses":{"200":{"description":"OK"}}}},"/ready":{"get":{"summary":"Readiness probe","tags":["core"],"responses":{"200":{"description":"OK"}}}},"/api/v1/admin/ping":{"get":{"summary":"Administrative ping","tags":["admin"],"security":[{"ApiKeyAuth":[]}],"responses":{"200":{"description":"OK"},"401":{"description":"Missing API key"},"403":{"description":"Invalid API key"}}}},"/ws":{"get":{"summary":"WebSocket endpoint","tags":["realtime"],"responses":{"101":{"description":"Switching Protocols"}}}}},"components":{"securitySchemes":{"ApiKeyAuth":{"type":"apiKey","in":"header","name":"X-API-Key"}}}}`)
}

// SwaggerHTML returns a Swagger UI page bound to the OpenAPI JSON URL.
func SwaggerHTML(openAPIPath string) string {
	return fmt.Sprintf(`<!doctype html><html><head><meta charset="utf-8"/><title>pixelsv Swagger</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>window.onload=function(){window.ui=SwaggerUIBundle({url:%q,dom_id:"#swagger-ui"});};</script></body></html>`, openAPIPath)
}
