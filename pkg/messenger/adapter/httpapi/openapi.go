package httpapi

// OpenAPIPaths returns OpenAPI path items owned by messenger HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	id := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	friendID := []map[string]any{
		{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
		{"name": "friendId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
	}
	errContent := errResponseContent()
	friendsListContent := map[string]any{"application/json": map[string]any{"schema": map[string]any{
		"type": "array", "items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"user_one_id": map[string]any{"type": "integer"},
				"user_two_id": map[string]any{"type": "integer"},
				"relationship": map[string]any{"type": "integer"},
				"created_at": map[string]any{"type": "string", "format": "date-time"},
			},
		},
	}}}
	requestsContent := map[string]any{"application/json": map[string]any{"schema": map[string]any{
		"type": "array", "items": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer"},
				"from_user_id": map[string]any{"type": "integer"},
				"to_user_id": map[string]any{"type": "integer"},
				"created_at": map[string]any{"type": "string", "format": "date-time"},
			},
		},
	}}}
	relContent := map[string]any{"application/json": map[string]any{"schema": map[string]any{
		"type": "object", "properties": map[string]any{"type": map[string]any{"type": "integer"}},
	}}}
	noContent := map[string]any{}
	return map[string]any{
		"/api/v1/users/{id}/friends": map[string]any{
			"get":  getFriendsOp(id, apiKey, friendsListContent, errContent),
			"post": addFriendOp(id, apiKey, noContent, errContent),
		},
		"/api/v1/users/{id}/friends/{friendId}": map[string]any{
			"delete": removeFriendOp(friendID, apiKey, noContent, errContent),
		},
		"/api/v1/users/{id}/friends/requests": map[string]any{
			"get": getRequestsOp(id, apiKey, requestsContent, errContent),
		},
		"/api/v1/users/{id}/friends/{friendId}/relationship": map[string]any{
			"get":   getRelationshipOp(friendID, apiKey, relContent, errContent),
			"patch": patchRelationshipOp(friendID, apiKey, relContent, errContent),
		},
	}
}

// errResponseContent returns a content block referencing the shared ErrorResponse schema.
func errResponseContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

// getFriendsOp returns the GET /friends operation map.
func getFriendsOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{"tags": []string{"messenger"}, "summary": "List user friends",
		"parameters": params, "security": sec,
		"responses": map[string]any{"200": map[string]any{"description": "Friend list", "content": ok}, "401": map[string]any{"description": "Unauthorized", "content": fail}}}
}

// addFriendOp returns the POST /friends operation map.
func addFriendOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"friend_id"}, "properties": map[string]any{"friend_id": map[string]any{"type": "integer"}}}}}}
	return map[string]any{"tags": []string{"messenger"}, "summary": "Force-add friendship",
		"parameters": params, "security": sec, "requestBody": body,
		"responses": map[string]any{"204": map[string]any{"description": "Created", "content": ok}, "400": map[string]any{"description": "Bad request", "content": fail}}}
}

// removeFriendOp returns the DELETE /friends/{friendId} operation map.
func removeFriendOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{"tags": []string{"messenger"}, "summary": "Remove friendship",
		"parameters": params, "security": sec,
		"responses": map[string]any{"204": map[string]any{"description": "Removed", "content": ok}, "404": map[string]any{"description": "Not found", "content": fail}}}
}

// getRequestsOp returns the GET /friends/requests operation map.
func getRequestsOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{"tags": []string{"messenger"}, "summary": "List pending friend requests",
		"parameters": params, "security": sec,
		"responses": map[string]any{"200": map[string]any{"description": "Requests", "content": ok}, "401": map[string]any{"description": "Unauthorized", "content": fail}}}
}

// getRelationshipOp returns the GET /relationship operation map.
func getRelationshipOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{"tags": []string{"messenger"}, "summary": "Get relationship type",
		"parameters": params, "security": sec,
		"responses": map[string]any{"200": map[string]any{"description": "Relationship type", "content": ok}, "404": map[string]any{"description": "Not found", "content": fail}}}
}

// patchRelationshipOp returns the PATCH /relationship operation map.
func patchRelationshipOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"type"}, "properties": map[string]any{"type": map[string]any{"type": "integer"}}}}}}
	return map[string]any{"tags": []string{"messenger"}, "summary": "Set relationship type",
		"parameters": params, "security": sec, "requestBody": body,
		"responses": map[string]any{"200": map[string]any{"description": "Updated", "content": ok}, "400": map[string]any{"description": "Bad request", "content": fail}}}
}
