package httpapi

// OpenAPIPaths returns OpenAPI path items owned by economy HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	idParam := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := ecoErrContent()
	return map[string]any{
		"/api/v1/marketplace/offers": map[string]any{
			"get":  listOffersOp(apiKey, errContent),
			"post": createOfferOp(apiKey, errContent),
		},
		"/api/v1/marketplace/offers/{id}": map[string]any{
			"get":    getOfferOp(idParam, apiKey, errContent),
			"delete": cancelOfferOp(idParam, apiKey, errContent),
		},
		"/api/v1/marketplace/history/{spriteId}": map[string]any{
			"get": getPriceHistoryOp(apiKey, errContent),
		},
		"/api/v1/marketplace/sellers/{sellerId}/offers": map[string]any{
			"get": listSellerOffersOp(apiKey, errContent),
		},
	}
}

func ecoErrContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func ecoJSON(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func marketplaceOfferSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":            map[string]any{"type": "integer"},
		"seller_id":     map[string]any{"type": "integer"},
		"item_id":       map[string]any{"type": "integer"},
		"definition_id": map[string]any{"type": "integer"},
		"asking_price":  map[string]any{"type": "integer"},
		"state":         map[string]any{"type": "string", "enum": []string{"open", "sold", "expired", "cancelled"}},
		"buyer_id":      map[string]any{"type": "integer", "nullable": true},
		"sold_at":       map[string]any{"type": "string", "format": "date-time", "nullable": true},
		"expire_at":     map[string]any{"type": "string", "format": "date-time"},
		"created_at":    map[string]any{"type": "string", "format": "date-time"},
	}}
}

func marketplaceOfferRequestSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"item_id", "asking_price"},
		"properties": map[string]any{
			"item_id":      map[string]any{"type": "integer"},
			"asking_price": map[string]any{"type": "integer", "minimum": 1},
		}}
}

func priceHistorySchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":          map[string]any{"type": "integer"},
		"sprite_id":   map[string]any{"type": "integer"},
		"day_offset":  map[string]any{"type": "integer", "description": "Days ago this entry represents"},
		"avg_price":   map[string]any{"type": "integer"},
		"sold_count":  map[string]any{"type": "integer"},
		"recorded_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func offersListSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"total":  map[string]any{"type": "integer"},
		"offset": map[string]any{"type": "integer"},
		"limit":  map[string]any{"type": "integer"},
		"items":  map[string]any{"type": "array", "items": marketplaceOfferSchema()},
	}}
}

func listOffersOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "List open offers", "security": apiKey,
		"parameters": []map[string]any{
			{"name": "min_price", "in": "query", "schema": map[string]any{"type": "integer"}},
			{"name": "max_price", "in": "query", "schema": map[string]any{"type": "integer"}},
			{"name": "offset", "in": "query", "schema": map[string]any{"type": "integer", "default": 0}},
			{"name": "limit", "in": "query", "schema": map[string]any{"type": "integer", "default": 50}},
		},
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": ecoJSON(offersListSchema())}}}
}

func createOfferOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "Create offer", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": ecoJSON(marketplaceOfferRequestSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": ecoJSON(marketplaceOfferSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func getOfferOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "Get offer", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": ecoJSON(marketplaceOfferSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func cancelOfferOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "Cancel offer", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func getPriceHistoryOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	spriteParam := []map[string]any{{"name": "spriteId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "Get price history", "security": apiKey, "parameters": spriteParam,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": ecoJSON(map[string]any{"type": "array", "items": priceHistorySchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func listSellerOffersOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	sellerParam := []map[string]any{{"name": "sellerId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{"tags": []string{"Marketplace"}, "summary": "List seller offers", "security": apiKey, "parameters": sellerParam,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": ecoJSON(map[string]any{"type": "array", "items": marketplaceOfferSchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}
