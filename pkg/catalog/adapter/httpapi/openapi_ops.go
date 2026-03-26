package httpapi

// listPagesOp returns the OpenAPI operation for GET /api/v1/catalog/pages.
func listPagesOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List pages", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": pageSchema()})}}}
}

// createPageOp returns the OpenAPI operation for POST /api/v1/catalog/pages.
func createPageOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Create page", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": json(pageRequestSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": json(pageSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

// getPageOp returns the OpenAPI operation for GET /api/v1/catalog/pages/{id}.
func getPageOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Get page", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": json(pageSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

// createOfferOp returns the OpenAPI operation for POST /api/v1/catalog/pages/{id}/offers.
func createOfferOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "Create offer", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": json(offerSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": json(offerSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

// listOffersOp returns the OpenAPI operation for GET /api/v1/catalog/pages/{id}/offers.
func listOffersOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List page offers", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": offerSchema()})},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

// listVouchersOp returns the OpenAPI operation for GET /api/v1/catalog/vouchers.
func listVouchersOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Catalog"}, "summary": "List vouchers", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": json(map[string]any{"type": "array", "items": voucherSchema()})}}}
}

// redeemVoucherOp returns the OpenAPI operation for POST /api/v1/catalog/vouchers/redeem.
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
