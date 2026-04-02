package httpapi

// OpenAPIPaths returns OpenAPI path items owned by navigator HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	idParam := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := navErrContent()
	return map[string]any{
		"/api/v1/navigator/categories": map[string]any{
			"get":  listCategoriesOp(apiKey),
			"post": createCategoryOp(apiKey, errContent),
		},
		"/api/v1/navigator/categories/{id}": map[string]any{
			"delete": deleteCategoryOp(idParam, apiKey, errContent),
		},
		"/api/v1/navigator/rooms": map[string]any{
			"get": listRoomsOp(apiKey),
		},
		"/api/v1/navigator/rooms/{id}": map[string]any{
			"get":    getRoomOp(idParam, apiKey, errContent),
			"delete": deleteRoomOp(idParam, apiKey, errContent),
		},
	}
}

func navErrContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func navJSON(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func categorySchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "caption": map[string]any{"type": "string"},
		"visible": map[string]any{"type": "boolean"}, "order_num": map[string]any{"type": "integer"},
		"icon_image": map[string]any{"type": "integer"}, "category_type": map[string]any{"type": "string"},
	}}
}

func roomSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id": map[string]any{"type": "integer"}, "name": map[string]any{"type": "string"},
		"owner_id": map[string]any{"type": "integer"}, "owner_name": map[string]any{"type": "string"},
		"state": map[string]any{"type": "string"}, "category_id": map[string]any{"type": "integer"},
		"max_users": map[string]any{"type": "integer"}, "score": map[string]any{"type": "integer"},
	}}
}

func listCategoriesOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "List categories", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": navJSON(map[string]any{"type": "array", "items": categorySchema()})}}}
}

func createCategoryOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "Create category", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": navJSON(categorySchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": navJSON(categorySchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func deleteCategoryOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "Delete category", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listRoomsOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "List rooms", "security": apiKey,
		"parameters": []map[string]any{
			{"name": "q", "in": "query", "schema": map[string]any{"type": "string"}},
			{"name": "offset", "in": "query", "schema": map[string]any{"type": "integer", "default": 0}},
			{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "default": 20}},
		},
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": navJSON(map[string]any{
			"type": "object", "properties": map[string]any{"rooms": map[string]any{"type": "array", "items": roomSchema()}, "total": map[string]any{"type": "integer"}},
		})}}}
}

func getRoomOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "Get room", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": navJSON(roomSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func deleteRoomOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Navigator"}, "summary": "Delete room", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}
