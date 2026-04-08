package httpapi

// OpenAPIPaths returns OpenAPI path items owned by moderation HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	errContent := map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
	return map[string]any{
		"/api/v1/moderation/actions": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Moderation"},
				"summary":     "List moderation actions",
				"operationId": "listModerationActions",
				"security":    apiKey,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Action list",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"type": "array", "items": actionSchema()},
							},
						},
					},
					"500": map[string]any{"description": "Internal error", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/alerts": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Moderation"},
				"summary":     "List moderation alerts registry",
				"operationId": "listModerationAlerts",
				"security":    apiKey,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Alert list",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"type": "array", "items": actionSchema()},
							},
						},
					},
					"500": map[string]any{"description": "Internal error", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/actions/{id}": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Moderation"},
				"summary":     "Get moderation action by ID",
				"operationId": "getModerationAction",
				"security":    apiKey,
				"parameters":  []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Action detail",
						"content": map[string]any{
							"application/json": map[string]any{"schema": actionSchema()},
						},
					},
					"404": map[string]any{"description": "Not found", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/users/{userId}/actions": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Moderation"},
				"summary":     "Get user moderation history",
				"operationId": "getUserModerationHistory",
				"security":    apiKey,
				"parameters":  []map[string]any{userIDParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "User action history",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"type": "array", "items": actionSchema()},
							},
						},
					},
					"400": map[string]any{"description": "Bad request", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/actions/{id}/deactivate": map[string]any{
			"patch": map[string]any{
				"tags":        []string{"Moderation"},
				"summary":     "Deactivate a moderation action",
				"operationId": "deactivateModerationAction",
				"security":    apiKey,
				"parameters":  []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Action deactivated",
						"content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
						},
					},
					"400": map[string]any{"description": "Bad request", "content": errContent},
				},
			},
		},
	}
}

// actionSchema returns the OpenAPI schema for a moderation action.
func actionSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id":               map[string]any{"type": "integer"},
			"scope":            map[string]any{"type": "string"},
			"action_type":      map[string]any{"type": "string"},
			"target_user_id":   map[string]any{"type": "integer"},
			"issuer_id":        map[string]any{"type": "integer"},
			"room_id":          map[string]any{"type": "integer"},
			"reason":           map[string]any{"type": "string"},
			"duration_minutes": map[string]any{"type": "integer"},
			"expires_at":       map[string]any{"type": "string", "format": "date-time"},
			"active":           map[string]any{"type": "boolean"},
			"created_at":       map[string]any{"type": "string", "format": "date-time"},
		},
	}
}

// idParam returns a path parameter schema for action ID.
func idParam() map[string]any {
	return map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}
}

// userIDParam returns a path parameter schema for user ID.
func userIDParam() map[string]any {
	return map[string]any{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}
}
