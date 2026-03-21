package httpapi

// OpenAPIPaths returns OpenAPI path items owned by inventory HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := invErrContent()
	return map[string]any{
		"/api/v1/inventory/{userId}/credits": map[string]any{
			"get": getCreditsOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/currencies": map[string]any{
			"get": listCurrenciesOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/badges": map[string]any{
			"get":  listBadgesOp(userParam, apiKey, errContent),
			"post": awardBadgeOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/effects": map[string]any{
			"get": listEffectsOp(userParam, apiKey, errContent),
		},
	}
}

func invErrContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func invJSON(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func currencySchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"user_id":    map[string]any{"type": "integer"},
		"type":       map[string]any{"type": "integer"},
		"amount":     map[string]any{"type": "integer"},
		"updated_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func badgeSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":         map[string]any{"type": "integer"},
		"user_id":    map[string]any{"type": "integer"},
		"badge_code": map[string]any{"type": "string"},
		"slot_id":    map[string]any{"type": "integer"},
		"created_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func effectSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":           map[string]any{"type": "integer"},
		"user_id":      map[string]any{"type": "integer"},
		"effect_id":    map[string]any{"type": "integer"},
		"duration":     map[string]any{"type": "integer", "description": "Total duration in seconds"},
		"quantity":     map[string]any{"type": "integer"},
		"activated_at": map[string]any{"type": "string", "format": "date-time", "nullable": true},
		"is_permanent": map[string]any{"type": "boolean"},
		"created_at":   map[string]any{"type": "string", "format": "date-time"},
	}}
}

func getCreditsOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	schema := map[string]any{"type": "object", "properties": map[string]any{
		"user_id": map[string]any{"type": "integer"},
		"credits": map[string]any{"type": "integer"},
	}}
	return map[string]any{"tags": []string{"Inventory"}, "summary": "Get credits balance", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(schema)},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func listCurrenciesOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "List currency balances", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(map[string]any{"type": "array", "items": currencySchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func listBadgesOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "List badges", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(map[string]any{"type": "array", "items": badgeSchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func awardBadgeOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	req := map[string]any{"type": "object", "required": []string{"badge_code"}, "properties": map[string]any{
		"badge_code": map[string]any{"type": "string"},
	}}
	return map[string]any{"tags": []string{"Inventory"}, "summary": "Award badge", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": invJSON(req)},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": invJSON(badgeSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
			"409": map[string]any{"description": "Conflict", "content": errContent},
		}}
}

func listEffectsOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "List effects", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(map[string]any{"type": "array", "items": effectSchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}
