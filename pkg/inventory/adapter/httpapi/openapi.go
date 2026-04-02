package httpapi

// OpenAPIPaths returns OpenAPI path items owned by inventory HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := invErrContent()
	return map[string]any{
		"/api/v1/inventory/{userId}/credits": map[string]any{
			"get":  getCreditsOp(userParam, apiKey, errContent),
			"post": addCreditsOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/currencies": map[string]any{
			"get": listCurrenciesOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/currencies/{type}": map[string]any{
			"post": addCurrencyOp(currencyTypeParams(), apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/badges": map[string]any{
			"get":  listBadgesOp(userParam, apiKey, errContent),
			"post": awardBadgeOp(userParam, apiKey, errContent),
		},
		"/api/v1/inventory/{userId}/badges/{code}": map[string]any{
			"delete": revokeBadgeOp(badgeCodeParams(), apiKey, errContent),
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

func currencyTypeParams() []map[string]any {
	return []map[string]any{
		{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
		{"name": "type", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 0}},
	}
}

func badgeCodeParams() []map[string]any {
	return []map[string]any{
		{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}},
		{"name": "code", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
	}
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

func modifyAmountSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"amount"}, "properties": map[string]any{
		"amount": map[string]any{"type": "integer", "description": "Amount to add"},
		"source": map[string]any{"type": "string", "description": "Optional audit source tag"},
	}}
}
