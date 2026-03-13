package httpapi

// OpenAPIPaths returns OpenAPI path items owned by user HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	id := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{
		"/api/v1/users/{id}": map[string]any{
			"get":   map[string]any{"tags": []string{"user"}, "summary": "Get user profile", "parameters": id, "responses": map[string]any{"200": map[string]any{"description": "User profile"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
			"patch": map[string]any{"tags": []string{"user"}, "summary": "Update user profile", "parameters": id, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"figure": map[string]any{"type": "string"}, "gender": map[string]any{"type": "string"}, "motto": map[string]any{"type": "string"}, "home_room_id": map[string]any{"type": "integer", "minimum": -1}}}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated user profile"}, "400": map[string]any{"description": "Invalid payload"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/settings": map[string]any{
			"get":   map[string]any{"tags": []string{"user"}, "summary": "Get user settings", "parameters": id, "responses": map[string]any{"200": map[string]any{"description": "User settings"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
			"patch": map[string]any{"tags": []string{"user"}, "summary": "Update user settings", "parameters": id, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object"}}}}, "responses": map[string]any{"200": map[string]any{"description": "Updated user settings"}, "400": map[string]any{"description": "Invalid payload"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/respect": map[string]any{
			"post": map[string]any{"tags": []string{"user"}, "summary": "Send user respect", "parameters": id, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"actor_user_id"}, "properties": map[string]any{"actor_user_id": map[string]any{"type": "integer", "minimum": 1}}}}}}, "responses": map[string]any{"200": map[string]any{"description": "Respect sent"}, "400": map[string]any{"description": "Invalid payload"}, "404": map[string]any{"description": "User not found"}, "409": map[string]any{"description": "Daily limit reached"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/wardrobe": map[string]any{
			"get": map[string]any{"tags": []string{"user"}, "summary": "Get user wardrobe", "parameters": id, "responses": map[string]any{"200": map[string]any{"description": "Wardrobe slots"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/respects": map[string]any{
			"get": map[string]any{"tags": []string{"user"}, "summary": "Get user respect history", "parameters": append(id, map[string]any{"name": "limit", "in": "query", "required": false, "schema": map[string]any{"type": "integer", "minimum": 1}}, map[string]any{"name": "offset", "in": "query", "required": false, "schema": map[string]any{"type": "integer", "minimum": 0}}), "responses": map[string]any{"200": map[string]any{"description": "Respect history"}, "404": map[string]any{"description": "User not found"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/name-change": map[string]any{
			"post": map[string]any{"tags": []string{"user"}, "summary": "Force user name change", "parameters": id, "requestBody": map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"name"}, "properties": map[string]any{"name": map[string]any{"type": "string"}}}}}}, "responses": map[string]any{"200": map[string]any{"description": "Name changed"}, "400": map[string]any{"description": "Invalid payload"}, "404": map[string]any{"description": "User not found"}, "409": map[string]any{"description": "Name change rejected"}, "401": map[string]any{"description": "Invalid API key"}}, "security": apiKey},
		},
	}
}
