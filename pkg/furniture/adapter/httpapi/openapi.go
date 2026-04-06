package httpapi

// OpenAPIPaths returns OpenAPI path items owned by furniture HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	idParam := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := errResponseContent()
	return map[string]any{
		"/api/v1/furniture/definitions": map[string]any{
			"get":  listDefsOp(apiKey, errContent),
			"post": createDefOp(apiKey, errContent),
		},
		"/api/v1/furniture/definitions/{id}": map[string]any{
			"get":    getDefOp(idParam, apiKey, errContent),
			"patch":  patchDefOp(idParam, apiKey, errContent),
			"delete": deleteDefOp(idParam, apiKey, errContent),
		},
		"/api/v1/furniture/items/user/{userId}": map[string]any{
			"get": listUserItemsOp(apiKey, errContent),
		},
		"/api/v1/furniture/items/{id}/transfer": map[string]any{
			"post": transferItemOp(idParam, apiKey, errContent),
		},
	}
}

func errResponseContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func jsonContent(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func definitionSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":                     map[string]any{"type": "integer"},
		"item_name":              map[string]any{"type": "string"},
		"public_name":            map[string]any{"type": "string"},
		"item_type":              map[string]any{"type": "string", "enum": []string{"s", "i", "e", "h", "v", "r", "b"}},
		"width":                  map[string]any{"type": "integer"},
		"length":                 map[string]any{"type": "integer"},
		"stack_height":           map[string]any{"type": "number", "format": "double"},
		"can_stack":              map[string]any{"type": "boolean"},
		"can_sit":                map[string]any{"type": "boolean"},
		"can_lay":                map[string]any{"type": "boolean"},
		"is_walkable":            map[string]any{"type": "boolean"},
		"sprite_id":              map[string]any{"type": "integer"},
		"allow_recycle":          map[string]any{"type": "boolean"},
		"allow_trade":            map[string]any{"type": "boolean"},
		"allow_marketplace_sell": map[string]any{"type": "boolean"},
		"allow_gift":             map[string]any{"type": "boolean"},
		"allow_inventory_stack":  map[string]any{"type": "boolean"},
		"interaction_type":       map[string]any{"type": "string"},
	}}
}

func definitionRequestSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"item_name", "item_type", "sprite_id"},
		"properties": map[string]any{
			"item_name":              map[string]any{"type": "string"},
			"public_name":            map[string]any{"type": "string"},
			"item_type":              map[string]any{"type": "string", "enum": []string{"s", "i", "e", "h", "v", "r", "b"}},
			"width":                  map[string]any{"type": "integer"},
			"length":                 map[string]any{"type": "integer"},
			"stack_height":           map[string]any{"type": "number"},
			"can_stack":              map[string]any{"type": "boolean"},
			"can_sit":                map[string]any{"type": "boolean"},
			"can_lay":                map[string]any{"type": "boolean"},
			"is_walkable":            map[string]any{"type": "boolean"},
			"sprite_id":              map[string]any{"type": "integer"},
			"allow_recycle":          map[string]any{"type": "boolean"},
			"allow_trade":            map[string]any{"type": "boolean"},
			"allow_marketplace_sell": map[string]any{"type": "boolean"},
			"allow_gift":             map[string]any{"type": "boolean"},
			"allow_inventory_stack":  map[string]any{"type": "boolean"},
			"interaction_type":       map[string]any{"type": "string"},
		}}
}

func itemSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":             map[string]any{"type": "integer"},
		"user_id":        map[string]any{"type": "integer"},
		"room_id":        map[string]any{"type": "integer"},
		"definition_id":  map[string]any{"type": "integer"},
		"extra_data":     map[string]any{"type": "string"},
		"limited_number": map[string]any{"type": "integer"},
		"limited_total":  map[string]any{"type": "integer"},
		"created_at":     map[string]any{"type": "string", "format": "date-time"},
	}}
}

func listDefsOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Furniture"}, "summary": "List definitions", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": jsonContent(map[string]any{"type": "array", "items": definitionSchema()})}}}
}

func createDefOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Furniture"}, "summary": "Create definition", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": jsonContent(definitionRequestSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": jsonContent(definitionSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func getDefOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Furniture"}, "summary": "Get definition", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": jsonContent(definitionSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func patchDefOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Furniture"}, "summary": "Update definition", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": jsonContent(definitionRequestSchema())},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": jsonContent(definitionSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func deleteDefOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Furniture"}, "summary": "Delete definition", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listUserItemsOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{"tags": []string{"Furniture"}, "summary": "List user items", "security": apiKey, "parameters": userParam,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": jsonContent(map[string]any{"type": "array", "items": itemSchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func transferItemOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	req := map[string]any{"type": "object", "required": []string{"new_user_id"}, "properties": map[string]any{
		"new_user_id": map[string]any{"type": "integer"},
	}}
	return map[string]any{"tags": []string{"Furniture"}, "summary": "Transfer item", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": jsonContent(req)},
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}
