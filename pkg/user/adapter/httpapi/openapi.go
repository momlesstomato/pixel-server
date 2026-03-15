package httpapi

// OpenAPIPaths returns OpenAPI path items owned by user HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	id := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := errResponseContent()
	profileContent := map[string]any{"application/json": map[string]any{"schema": userProfileSchema()}}
	settingsContent := map[string]any{"application/json": map[string]any{"schema": userSettingsSchema()}}
	return map[string]any{
		"/api/v1/users/{id}": map[string]any{
			"get":   profileGetOp(id, apiKey, profileContent, errContent),
			"patch": profilePatchOp(id, apiKey, profileContent, errContent),
		},
		"/api/v1/users/{id}/settings": map[string]any{
			"get":   settingsGetOp(id, apiKey, settingsContent, errContent),
			"patch": settingsPatchOp(id, apiKey, settingsContent, errContent),
		},
		"/api/v1/users/{id}/respect": map[string]any{
			"post": respectPostOp(id, apiKey, errContent),
		},
		"/api/v1/users/{id}/wardrobe": map[string]any{
			"get": wardrobeGetOp(id, apiKey, errContent),
		},
		"/api/v1/users/{id}/respects": map[string]any{
			"get": respectsGetOp(id, apiKey, errContent),
		},
		"/api/v1/users/{id}/name-change": map[string]any{
			"post": nameChangePostOp(id, apiKey, errContent),
		},
	}
}

// errResponseContent returns a content block referencing the shared ErrorResponse schema.
func errResponseContent() map[string]any {
	return map[string]any{
		"application/json": map[string]any{
			"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
		},
	}
}

// profileGetOp returns the GET /api/v1/users/{id} operation map.
func profileGetOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{
		"tags": []string{"user"}, "summary": "Get user profile",
		"parameters": params,
		"responses":  map[string]any{"200": map[string]any{"description": "User profile", "content": ok}, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":   sec,
	}
}

// profilePatchOp returns the PATCH /api/v1/users/{id} operation map.
func profilePatchOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"figure": map[string]any{"type": "string"}, "gender": map[string]any{"type": "string"}, "motto": map[string]any{"type": "string"}, "home_room_id": map[string]any{"type": "integer", "minimum": -1}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Update user profile",
		"parameters": params, "requestBody": body,
		"responses": map[string]any{"200": map[string]any{"description": "Updated", "content": ok}, "400": map[string]any{"description": "Invalid payload", "content": fail}, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":  sec,
	}
}

// settingsGetOp returns the GET /api/v1/users/{id}/settings operation map.
func settingsGetOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	return map[string]any{
		"tags": []string{"user"}, "summary": "Get user settings",
		"parameters": params,
		"responses":  map[string]any{"200": map[string]any{"description": "User settings", "content": ok}, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":   sec,
	}
}

// settingsPatchOp returns the PATCH /api/v1/users/{id}/settings operation map.
func settingsPatchOp(params, sec []map[string]any, ok, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"volume_system": map[string]any{"type": "integer"}, "volume_furni": map[string]any{"type": "integer"}, "volume_trax": map[string]any{"type": "integer"}, "old_chat": map[string]any{"type": "boolean"}, "room_invites": map[string]any{"type": "boolean"}, "camera_follow": map[string]any{"type": "boolean"}, "flags": map[string]any{"type": "integer"}, "chat_type": map[string]any{"type": "integer"}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Update user settings",
		"parameters": params, "requestBody": body,
		"responses": map[string]any{"200": map[string]any{"description": "Updated", "content": ok}, "400": map[string]any{"description": "Invalid payload", "content": fail}, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":  sec,
	}
}

// respectPostOp returns the POST /api/v1/users/{id}/respect operation map.
func respectPostOp(params, sec []map[string]any, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"actor_user_id"}, "properties": map[string]any{"actor_user_id": map[string]any{"type": "integer", "minimum": 1}}}}}}
	respects200 := map[string]any{"description": "Respect sent", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"respects_received": map[string]any{"type": "integer"}, "remaining": map[string]any{"type": "integer"}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Send user respect",
		"parameters": params, "requestBody": body,
		"responses": map[string]any{"200": respects200, "400": map[string]any{"description": "Invalid payload", "content": fail}, "404": map[string]any{"description": "User not found", "content": fail}, "409": map[string]any{"description": "Daily limit reached", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":  sec,
	}
}

// wardrobeGetOp returns the GET /api/v1/users/{id}/wardrobe operation map.
func wardrobeGetOp(params, sec []map[string]any, fail map[string]any) map[string]any {
	slots200 := map[string]any{"description": "Wardrobe slots", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"slots": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{"slot_id": map[string]any{"type": "integer"}, "figure": map[string]any{"type": "string"}, "gender": map[string]any{"type": "string"}}}}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Get user wardrobe",
		"parameters": params,
		"responses":  map[string]any{"200": slots200, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":   sec,
	}
}

// respectsGetOp returns the GET /api/v1/users/{id}/respects operation map.
func respectsGetOp(params, sec []map[string]any, fail map[string]any) map[string]any {
	extraParams := append(params, map[string]any{"name": "limit", "in": "query", "required": false, "schema": map[string]any{"type": "integer", "minimum": 1}}, map[string]any{"name": "offset", "in": "query", "required": false, "schema": map[string]any{"type": "integer", "minimum": 0}})
	records200 := map[string]any{"description": "Respect history", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"records": map[string]any{"type": "array"}, "limit": map[string]any{"type": "integer"}, "offset": map[string]any{"type": "integer"}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Get user respect history",
		"parameters": extraParams,
		"responses":  map[string]any{"200": records200, "404": map[string]any{"description": "User not found", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":   sec,
	}
}

// nameChangePostOp returns the POST /api/v1/users/{id}/name-change operation map.
func nameChangePostOp(params, sec []map[string]any, fail map[string]any) map[string]any {
	body := map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"name"}, "properties": map[string]any{"name": map[string]any{"type": "string"}}}}}}
	name200 := map[string]any{"description": "Name changed", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"result_code": map[string]any{"type": "integer"}, "name": map[string]any{"type": "string"}, "suggestions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}}}}
	return map[string]any{
		"tags": []string{"user"}, "summary": "Force user name change",
		"parameters": params, "requestBody": body,
		"responses": map[string]any{"200": name200, "400": map[string]any{"description": "Invalid payload", "content": fail}, "404": map[string]any{"description": "User not found", "content": fail}, "409": map[string]any{"description": "Name change rejected", "content": fail}, "401": map[string]any{"description": "Unauthorized", "content": fail}},
		"security":  sec,
	}
}

// userProfileSchema returns the JSON Schema for a user profile response.
func userProfileSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{"id", "username", "figure", "gender", "motto"},
		"properties": map[string]any{
			"id":                map[string]any{"type": "integer"},
			"username":          map[string]any{"type": "string"},
			"figure":            map[string]any{"type": "string"},
			"gender":            map[string]any{"type": "string"},
			"motto":             map[string]any{"type": "string"},
			"real_name":         map[string]any{"type": "string"},
			"respects_received": map[string]any{"type": "integer"},
			"home_room_id":      map[string]any{"type": "integer"},
			"can_change_name":   map[string]any{"type": "boolean"},
			"noobness_level":    map[string]any{"type": "integer"},
			"safety_locked":     map[string]any{"type": "boolean"},
			"group_id":          map[string]any{"type": "integer"},
		},
	}
}

// userSettingsSchema returns the JSON Schema for a user settings response.
func userSettingsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"VolumeSystem": map[string]any{"type": "integer"},
			"VolumeFurni":  map[string]any{"type": "integer"},
			"VolumeTrax":   map[string]any{"type": "integer"},
			"OldChat":      map[string]any{"type": "boolean"},
			"RoomInvites":  map[string]any{"type": "boolean"},
			"CameraFollow": map[string]any{"type": "boolean"},
			"Flags":        map[string]any{"type": "integer"},
			"ChatType":     map[string]any{"type": "integer"},
		},
	}
}

