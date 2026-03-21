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
			"get": listOffersOp(idParam, apiKey, errContent),
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
		"id":            map[string]any{"type": "integer"},
		"parent_id":     map[string]any{"type": "integer"},
		"caption":       map[string]any{"type": "string"},
		"icon_image":    map[string]any{"type": "integer"},
		"page_layout":   map[string]any{"type": "string"},
		"visible":       map[string]any{"type": "boolean"},
		"enabled":       map[string]any{"type": "boolean"},
		"min_rank":      map[string]any{"type": "integer"},
		"min_club_level": map[string]any{"type": "integer"},
		"order_num":     map[string]any{"type": "integer"},
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
			"min_rank":       map[string]any{"type": "integer"},
			"min_club_level": map[string]any{"type": "integer"},
			"order_num":      map[string]any{"type": "integer"},
		}}
}

func offerSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":                  map[string]any{"type": "integer"},
		"page_id":             map[string]any{"type": "integer"},
		"item_definition_id":  map[string]any{"type": "integer"},
		"catalog_name":        map[string]any{"type": "string"},
		"cost_primary":        map[string]any{"type": "integer"},
		"cost_primary_type":   map[string]any{"type": "integer"},
		"cost_secondary":      map[string]any{"type": "integer"},
		"cost_secondary_type": map[string]any{"type": "integer"},
		"amount":              map[string]any{"type": "integer"},
		"limited_total":       map[string]any{"type": "integer"},
		"limited_sells":       map[string]any{"type": "integer"},
		"offer_active":        map[string]any{"type": "boolean"},
		"extra_data":          map[string]any{"type": "string"},
		"badge_id":            map[string]any{"type": "string"},
		"club_only":           map[string]any{"type": "boolean"},
		"order_num":           map[string]any{"type": "integer"},
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

func listPagesOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List pages", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": pageSchema()})}}}
}

func createPageOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Create page", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": json(pageRequestSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": json(pageSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func getPageOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Get page", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": json(pageSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listOffersOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List page offers", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": offerSchema()})},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listVouchersOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List vouchers", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": voucherSchema()})}}}
}

func redeemVoucherOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	req := map[string]any{"type": "object", "required": []string{"code", "user_id"}, "properties": map[string]any{
		"code":    map[string]any{"type": "string"},
		"user_id": map[string]any{"type": "integer"},
	}}
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Redeem voucher", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": json(req)},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": json(voucherSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
			"409": map[string]any{"description": "Conflict", "content": errContent},
		}}
}
