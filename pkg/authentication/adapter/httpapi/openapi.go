package httpapi

// OpenAPIPaths returns OpenAPI path items owned by authentication HTTP routes.
func OpenAPIPaths() map[string]any {
	errContent := errResponseContent()
	return map[string]any{
		"/api/v1/sso": map[string]any{
			"post": map[string]any{
				"tags":        []string{"authentication"},
				"summary":     "Issue SSO ticket",
				"description": "Issues one single-use ticket bound to a user identifier.",
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type":     "object",
								"required": []string{"user_id"},
								"properties": map[string]any{
									"user_id":     map[string]any{"type": "integer", "minimum": 1},
									"ttl_seconds": map[string]any{"type": "integer", "minimum": 1},
								},
							},
						},
					},
				},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Ticket issued",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type":     "object",
									"required": []string{"ticket", "expires_at"},
									"properties": map[string]any{
										"ticket":     map[string]any{"type": "string"},
										"expires_at": map[string]any{"type": "string", "format": "date-time"},
									},
								},
							},
						},
					},
					"400": map[string]any{"description": "Invalid payload", "content": errContent},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
					"500": map[string]any{"description": "Internal server error", "content": errContent},
				},
				"security": []map[string]any{{"ApiKeyAuth": []string{}}},
			},
		},
	}
}

// errResponseContent returns an application/json content block referencing ErrorResponse schema.
func errResponseContent() map[string]any {
	return map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
}
