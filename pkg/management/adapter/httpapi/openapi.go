package httpapi

// OpenAPIPaths returns OpenAPI path items owned by management HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	errContent := map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
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
					"200": map[string]any{
						"description": "Session list",
						"content": map[string]any{
							"application/json": map[string]any{"schema": sessionListSchema()},
						},
					},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
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
					"200": map[string]any{
						"description": "Session details",
						"content": map[string]any{
							"application/json": map[string]any{"schema": sessionSchema()},
						},
					},
					"404": map[string]any{"description": "Session not found", "content": errContent},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
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
					"200": map[string]any{
						"description": "Session disconnected",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type":       "object",
									"properties": map[string]any{"disconnected": map[string]any{"type": "string"}},
								},
							},
						},
					},
					"404": map[string]any{"description": "Session not found", "content": errContent},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
				},
				"security": apiKey,
			},
		},
		"/api/v1/hotel/status": map[string]any{
			"get": map[string]any{
				"tags":    []string{"management"},
				"summary": "Get hotel status",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Hotel status",
						"content": map[string]any{
							"application/json": map[string]any{"schema": hotelStatusSchema()},
						},
					},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
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
					"200": map[string]any{
						"description": "Close scheduled",
						"content": map[string]any{
							"application/json": map[string]any{"schema": hotelStatusSchema()},
						},
					},
					"400": map[string]any{"description": "Invalid payload", "content": errContent},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
					"409": map[string]any{"description": "State transition conflict", "content": errContent},
				},
				"security": apiKey,
			},
		},
		"/api/v1/hotel/reopen": map[string]any{
			"post": map[string]any{
				"tags":    []string{"management"},
				"summary": "Reopen hotel",
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Hotel reopened",
						"content": map[string]any{
							"application/json": map[string]any{"schema": hotelStatusSchema()},
						},
					},
					"401": map[string]any{"description": "Invalid API key", "content": errContent},
					"409": map[string]any{"description": "State transition conflict", "content": errContent},
				},
				"security": apiKey,
			},
		},
	}
}
