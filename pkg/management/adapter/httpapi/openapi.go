package httpapi

// OpenAPIPaths returns OpenAPI path items owned by management HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	return map[string]any{
		"/api/v1/sessions": map[string]any{
			"get": map[string]any{
				"tags":        []string{"management"},
				"summary":     "List active sessions",
				"description": "Returns all active sessions. Optionally filter by instance query parameter.",
				"parameters": []map[string]any{
					{"name": "instance", "in": "query", "required": false, "schema": map[string]any{"type": "string"}},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "Session list"},
					"401": map[string]any{"description": "Invalid API key"},
				},
				"security": apiKey,
			},
		},
		"/api/v1/sessions/{connID}": map[string]any{
			"get": map[string]any{
				"tags":    []string{"management"},
				"summary": "Get session by connection ID",
				"parameters": []map[string]any{
					{"name": "connID", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "Session details"},
					"404": map[string]any{"description": "Session not found"},
					"401": map[string]any{"description": "Invalid API key"},
				},
				"security": apiKey,
			},
			"delete": map[string]any{
				"tags":    []string{"management"},
				"summary": "Disconnect session",
				"parameters": []map[string]any{
					{"name": "connID", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "Session disconnected"},
					"404": map[string]any{"description": "Session not found"},
					"401": map[string]any{"description": "Invalid API key"},
				},
				"security": apiKey,
			},
		},
		"/api/v1/hotel/status": map[string]any{
			"get": map[string]any{
				"tags":    []string{"management"},
				"summary": "Get hotel status",
				"responses": map[string]any{
					"200": map[string]any{"description": "Hotel status"},
					"401": map[string]any{"description": "Invalid API key"},
				},
				"security": apiKey,
			},
		},
		"/api/v1/hotel/close": map[string]any{
			"post": map[string]any{
				"tags":    []string{"management"},
				"summary": "Schedule hotel close",
				"requestBody": map[string]any{
					"required": true,
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"minutes_until_close": map[string]any{"type": "integer", "minimum": 0},
									"duration_minutes":    map[string]any{"type": "integer", "minimum": 1},
									"throw_users":         map[string]any{"type": "boolean"},
								},
							},
						},
					},
				},
				"responses": map[string]any{
					"200": map[string]any{"description": "Close scheduled"},
					"400": map[string]any{"description": "Invalid payload"},
					"401": map[string]any{"description": "Invalid API key"},
					"409": map[string]any{"description": "State transition conflict"},
				},
				"security": apiKey,
			},
		},
		"/api/v1/hotel/reopen": map[string]any{
			"post": map[string]any{
				"tags":    []string{"management"},
				"summary": "Reopen hotel",
				"responses": map[string]any{
					"200": map[string]any{"description": "Hotel reopened"},
					"401": map[string]any{"description": "Invalid API key"},
					"409": map[string]any{"description": "State transition conflict"},
				},
				"security": apiKey,
			},
		},
	}
}
