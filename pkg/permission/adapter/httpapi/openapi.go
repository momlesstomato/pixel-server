package httpapi

// OpenAPIPaths returns OpenAPI path items owned by permission HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	groupID := map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}
	permission := map[string]any{"name": "permission", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}
	return map[string]any{
		"/api/v1/groups": map[string]any{
			"get":  map[string]any{"tags": []string{"permission"}, "summary": "List permission groups", "responses": map[string]any{"200": map[string]any{"description": "Group list"}}, "security": apiKey},
			"post": map[string]any{"tags": []string{"permission"}, "summary": "Create permission group", "responses": map[string]any{"201": map[string]any{"description": "Group created"}}, "security": apiKey},
		},
		"/api/v1/groups/{id}": map[string]any{
			"get":    map[string]any{"tags": []string{"permission"}, "summary": "Get permission group", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group details"}}, "security": apiKey},
			"patch":  map[string]any{"tags": []string{"permission"}, "summary": "Update permission group", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group updated"}}, "security": apiKey},
			"delete": map[string]any{"tags": []string{"permission"}, "summary": "Delete permission group", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group deleted"}}, "security": apiKey},
		},
		"/api/v1/groups/{id}/permissions": map[string]any{
			"get":  map[string]any{"tags": []string{"permission"}, "summary": "List group permissions", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Permissions"}}, "security": apiKey},
			"post": map[string]any{"tags": []string{"permission"}, "summary": "Add group permissions", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Permissions updated"}}, "security": apiKey},
		},
		"/api/v1/groups/{id}/permissions/{permission}": map[string]any{
			"delete": map[string]any{"tags": []string{"permission"}, "summary": "Remove group permission", "parameters": []map[string]any{groupID, permission}, "responses": map[string]any{"200": map[string]any{"description": "Permission removed"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/group": map[string]any{
			"patch": map[string]any{"tags": []string{"permission"}, "summary": "Assign single group to user", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group assignment updated"}}, "security": apiKey},
		},
		"/api/v1/users/{id}/groups": map[string]any{
			"patch": map[string]any{"tags": []string{"permission"}, "summary": "Replace user groups", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group assignments updated"}}, "security": apiKey},
		},
	}
}
