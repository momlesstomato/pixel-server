package openapi

// ErrorResponseSchema returns the JSON Schema for the common error response body.
func ErrorResponseSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"error"},
		"properties": map[string]any{
			"error": map[string]any{
				"type":        "string",
				"description": "Human-readable error message.",
			},
		},
	}
}

// BuildComponents returns the OpenAPI components object including security schemes
// and shared named schemas.
func BuildComponents() map[string]any {
	return map[string]any{
		"securitySchemes": map[string]any{
			"ApiKeyAuth": map[string]any{
				"type": "apiKey", "in": "header", "name": "X-API-Key",
			},
		},
		"schemas": map[string]any{
			"ErrorResponse": ErrorResponseSchema(),
		},
	}
}
