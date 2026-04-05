package httpapi

// OpenAPIPaths returns OpenAPI path items owned by room HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	errContent := map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
	roomParam := []map[string]any{
		{"name": "roomId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
	}
	chatLogParams := append(roomParam,
		map[string]any{"name": "from", "in": "query", "required": false, "schema": map[string]any{"type": "string", "format": "date"}},
		map[string]any{"name": "to", "in": "query", "required": false, "schema": map[string]any{"type": "string", "format": "date"}},
	)
	return map[string]any{
		"/api/v1/rooms/{roomId}/chat-logs": map[string]any{
			"get": map[string]any{
				"tags":        []string{"Room"},
				"summary":     "List room chat logs filtered by date range",
				"operationId": "listRoomChatLogs",
				"security":    apiKey,
				"parameters":  chatLogParams,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Chat log entries",
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": chatLogArraySchema(),
							},
						},
					},
					"400": map[string]any{"description": "Bad request", "content": errContent},
					"500": map[string]any{"description": "Internal error", "content": errContent},
				},
			},
		},
	}
}

// chatLogArraySchema returns the OpenAPI schema for a chat log entry array.
func chatLogArraySchema() map[string]any {
	return map[string]any{
		"type": "array",
		"items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"room_id":    map[string]any{"type": "integer"},
				"user_id":    map[string]any{"type": "integer"},
				"username":   map[string]any{"type": "string"},
				"message":    map[string]any{"type": "string"},
				"chat_type":  map[string]any{"type": "string"},
				"created_at": map[string]any{"type": "string", "format": "date-time"},
			},
		},
	}
}
