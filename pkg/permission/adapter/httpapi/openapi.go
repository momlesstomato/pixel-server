package httpapi

// OpenAPIPaths returns OpenAPI path items owned by permission HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	errContent := errResponseContent()
	groupID := map[string]any{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}
	permission := map[string]any{"name": "permission", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}
	groupContent := map[string]any{"application/json": map[string]any{"schema": groupDetailsSchema()}}
	groupListContent := map[string]any{"application/json": map[string]any{"schema": groupListSchema()}}
	permListContent := map[string]any{"application/json": map[string]any{"schema": permissionListSchema()}}
	accessContent := map[string]any{"application/json": map[string]any{"schema": userAccessSchema()}}
	return map[string]any{
		"/api/v1/groups": map[string]any{
			"get":  map[string]any{"tags": []string{"permission"}, "summary": "List permission groups", "responses": map[string]any{"200": map[string]any{"description": "Group list", "content": groupListContent}}, "security": apiKey},
			"post": map[string]any{"tags": []string{"permission"}, "summary": "Create permission group", "requestBody": createGroupBody(), "responses": map[string]any{"201": map[string]any{"description": "Group created", "content": groupContent}, "400": map[string]any{"description": "Invalid payload", "content": errContent}, "401": map[string]any{"description": "Unauthorized", "content": errContent}}, "security": apiKey},
		},
		"/api/v1/groups/{id}": map[string]any{
			"get":    map[string]any{"tags": []string{"permission"}, "summary": "Get permission group", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group details", "content": groupContent}, "404": map[string]any{"description": "Not found", "content": errContent}}, "security": apiKey},
			"patch":  map[string]any{"tags": []string{"permission"}, "summary": "Update permission group", "parameters": []map[string]any{groupID}, "requestBody": patchGroupBody(), "responses": map[string]any{"200": map[string]any{"description": "Group updated", "content": groupContent}, "404": map[string]any{"description": "Not found", "content": errContent}}, "security": apiKey},
			"delete": map[string]any{"tags": []string{"permission"}, "summary": "Delete permission group", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Group deleted", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"deleted": map[string]any{"type": "integer"}}}}}}, "404": map[string]any{"description": "Not found", "content": errContent}, "409": map[string]any{"description": "Conflict", "content": errContent}}, "security": apiKey},
		},
		"/api/v1/groups/{id}/permissions": map[string]any{
			"get":  map[string]any{"tags": []string{"permission"}, "summary": "List group permissions", "parameters": []map[string]any{groupID}, "responses": map[string]any{"200": map[string]any{"description": "Permissions", "content": permListContent}, "404": map[string]any{"description": "Not found", "content": errContent}}, "security": apiKey},
			"post": map[string]any{"tags": []string{"permission"}, "summary": "Add group permissions", "parameters": []map[string]any{groupID}, "requestBody": addPermissionsBody(), "responses": map[string]any{"200": map[string]any{"description": "Permissions updated", "content": groupContent}, "404": map[string]any{"description": "Not found", "content": errContent}}, "security": apiKey},
		},
		"/api/v1/groups/{id}/permissions/{permission}": map[string]any{
			"delete": map[string]any{"tags": []string{"permission"}, "summary": "Remove group permission", "parameters": []map[string]any{groupID, permission}, "responses": map[string]any{"200": map[string]any{"description": "Permission removed", "content": groupContent}, "404": map[string]any{"description": "Not found", "content": errContent}}, "security": apiKey},
		},
		"/api/v1/users/{id}/group": map[string]any{
			"patch": map[string]any{"tags": []string{"permission"}, "summary": "Assign single group to user", "parameters": []map[string]any{groupID}, "requestBody": singleGroupBody(), "responses": map[string]any{"200": map[string]any{"description": "Group assignment updated", "content": accessContent}, "400": map[string]any{"description": "Invalid payload", "content": errContent}, "401": map[string]any{"description": "Unauthorized", "content": errContent}}, "security": apiKey},
		},
		"/api/v1/users/{id}/groups": map[string]any{
			"patch": map[string]any{"tags": []string{"permission"}, "summary": "Replace user groups", "parameters": []map[string]any{groupID}, "requestBody": multiGroupBody(), "responses": map[string]any{"200": map[string]any{"description": "Group assignments updated", "content": accessContent}, "400": map[string]any{"description": "Invalid payload", "content": errContent}, "401": map[string]any{"description": "Unauthorized", "content": errContent}}, "security": apiKey},
		},
	}
}

// errResponseContent returns a content block referencing the shared ErrorResponse schema.
func errResponseContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}
