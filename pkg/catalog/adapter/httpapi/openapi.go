package httpapi

// OpenAPIPaths returns OpenAPI path items owned by catalog HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	idParam := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := errContent()
	return map[string]any{
		"/api/v1/catalog/pages": map[string]any{
			"get":  listPagesOp(apiKey, errContent),
			"post": createPageOp(apiKey, errContent),
		},
		"/api/v1/catalog/pages/{id}": map[string]any{
			"get": getPageOp(idParam, apiKey, errContent),
		},
		"/api/v1/catalog/pages/{id}/offers": map[string]any{
			"get":  listOffersOp(idParam, apiKey, errContent),
			"post": createOfferOp(idParam, apiKey, errContent),
		},
		"/api/v1/catalog/vouchers": map[string]any{
			"get": listVouchersOp(apiKey),
		},
		"/api/v1/catalog/vouchers/redeem": map[string]any{
			"post": redeemVoucherOp(apiKey, errContent),
		},
	}
}

func errContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func pageSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":             map[string]any{"type": "integer"},
		"parent_id":      map[string]any{"type": "integer"},
		"caption":        map[string]any{"type": "string"},
		"icon_image":     map[string]any{"type": "integer"},
		"page_layout":    map[string]any{"type": "string"},
		"visible":        map[string]any{"type": "boolean"},
		"enabled":        map[string]any{"type": "boolean"},
		"min_permission": map[string]any{"type": "string", "description": "Dotted permission required to view page; empty means everyone"},
		"club_only":      map[string]any{"type": "boolean"},
		"order_num":      map[string]any{"type": "integer"},
	}}
}

func pageRequestSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"caption", "page_layout"},
		"properties": map[string]any{
			"parent_id":      map[string]any{"type": "integer"},
			"caption":        map[string]any{"type": "string"},
			"icon_image":     map[string]any{"type": "integer"},
			"page_layout":    map[string]any{"type": "string"},
			"visible":        map[string]any{"type": "boolean"},
			"enabled":        map[string]any{"type": "boolean"},
			"min_permission": map[string]any{"type": "string"},
			"club_only":      map[string]any{"type": "boolean"},
			"order_num":      map[string]any{"type": "integer"},
		}}
}

func offerSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":                   map[string]any{"type": "integer"},
		"page_id":              map[string]any{"type": "integer"},
		"item_definition_id":   map[string]any{"type": "integer"},
		"catalog_name":         map[string]any{"type": "string", "description": "Resolved from item_definitions.public_name; read-only"},
		"cost_credits":         map[string]any{"type": "integer", "description": "Credits price; zero for activity-point-only offers"},
		"cost_activity_points": map[string]any{"type": "integer", "description": "Activity-point price; zero when no activity-point charge"},
		"activity_point_type":  map[string]any{"type": "integer", "description": "Registered activity-point type ID (see currency_types)"},
		"amount":               map[string]any{"type": "integer"},
		"limited_total":        map[string]any{"type": "integer"},
		"limited_sells":        map[string]any{"type": "integer"},
		"offer_active":         map[string]any{"type": "boolean"},
		"extra_data":           map[string]any{"type": "string"},
		"badge_id":             map[string]any{"type": "string"},
		"club_only":            map[string]any{"type": "boolean"},
		"order_num":            map[string]any{"type": "integer"},
	}}
}

func voucherSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":                   map[string]any{"type": "integer"},
		"code":                 map[string]any{"type": "string"},
		"reward_type":          map[string]any{"type": "string", "enum": []string{"currency", "badge", "furniture"}},
		"reward_currency_type": map[string]any{"type": "integer", "nullable": true},
		"reward_data":          map[string]any{"type": "string"},
		"max_uses":             map[string]any{"type": "integer"},
		"current_uses":         map[string]any{"type": "integer"},
		"enabled":              map[string]any{"type": "boolean"},
		"created_at":           map[string]any{"type": "string", "format": "date-time"},
	}}
}

func json(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}
