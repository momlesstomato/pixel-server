package httpapi

// Phase2OpenAPIPaths returns OpenAPI path items for Phase 2 moderation routes.
func Phase2OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	errContent := map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
	return map[string]any{
		"/api/v1/moderation/tickets": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "List support tickets",
				"operationId": "listModerationTickets", "security": apiKey,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Ticket list", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": ticketSchema()}},
						},
					},
					"500": map[string]any{"description": "Internal error", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/tickets/{id}": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "Get ticket by ID",
				"operationId": "getModerationTicket", "security": apiKey,
				"parameters": []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Ticket detail", "content": map[string]any{
							"application/json": map[string]any{"schema": ticketSchema()},
						},
					},
					"404": map[string]any{"description": "Not found", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/tickets/{id}/close": map[string]any{
			"patch": map[string]any{
				"tags": []string{"Moderation"}, "summary": "Close a support ticket",
				"operationId": "closeModerationTicket", "security": apiKey,
				"parameters": []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Ticket closed", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
						},
					},
					"400": map[string]any{"description": "Bad request", "content": errContent},
				},
			},
		},
		"/api/v1/moderation/wordfilters": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "List word filters",
				"operationId": "listWordFilters", "security": apiKey,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Filter list", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": filterSchema()}},
						},
					},
				},
			},
		},
		"/api/v1/moderation/wordfilters/{id}": map[string]any{
			"delete": map[string]any{
				"tags": []string{"Moderation"}, "summary": "Delete word filter",
				"operationId": "deleteWordFilter", "security": apiKey,
				"parameters": []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Deleted", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
						},
					},
				},
			},
		},
		"/api/v1/moderation/presets": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "List moderation presets",
				"operationId": "listModerationPresets", "security": apiKey,
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Preset list", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": presetSchema()}},
						},
					},
				},
			},
		},
		"/api/v1/moderation/presets/{id}": map[string]any{
			"delete": map[string]any{
				"tags": []string{"Moderation"}, "summary": "Delete moderation preset",
				"operationId": "deleteModerationPreset", "security": apiKey,
				"parameters": []map[string]any{idParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Deleted", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "object"}},
						},
					},
				},
			},
		},
		"/api/v1/moderation/visits/users/{userId}": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "User room visit history",
				"operationId": "getUserVisits", "security": apiKey,
				"parameters": []map[string]any{userIDParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Visit list", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": visitSchema()}},
						},
					},
				},
			},
		},
		"/api/v1/moderation/visits/rooms/{roomId}": map[string]any{
			"get": map[string]any{
				"tags": []string{"Moderation"}, "summary": "Room visit history",
				"operationId": "getRoomVisits", "security": apiKey,
				"parameters": []map[string]any{roomIDParam()},
				"responses": map[string]any{
					"200": map[string]any{
						"description": "Visit list", "content": map[string]any{
							"application/json": map[string]any{"schema": map[string]any{"type": "array", "items": visitSchema()}},
						},
					},
				},
			},
		},
	}
}

func ticketSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "reporter_id": map[string]any{"type": "integer"},
		"reported_id": map[string]any{"type": "integer"}, "room_id": map[string]any{"type": "integer"},
		"category": map[string]any{"type": "string"}, "message": map[string]any{"type": "string"},
		"status": map[string]any{"type": "string"}, "created_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func filterSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "pattern": map[string]any{"type": "string"},
		"replacement": map[string]any{"type": "string"}, "is_regex": map[string]any{"type": "boolean"},
		"scope": map[string]any{"type": "string"}, "active": map[string]any{"type": "boolean"},
	}}
}

func presetSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "category": map[string]any{"type": "string"},
		"name": map[string]any{"type": "string"}, "action_type": map[string]any{"type": "string"},
		"default_duration": map[string]any{"type": "integer"}, "active": map[string]any{"type": "boolean"},
	}}
}

func visitSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "user_id": map[string]any{"type": "integer"},
		"room_id": map[string]any{"type": "integer"}, "visited_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

// roomIDParam returns a path parameter schema for room ID.
func roomIDParam() map[string]any {
	return map[string]any{"name": "roomId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}
}
