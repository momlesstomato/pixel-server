package httpapi

// OpenAPIPaths returns OpenAPI path items owned by subscription HTTP routes.
func OpenAPIPaths() map[string]any {
	apiKey := []map[string]any{{"ApiKeyAuth": []string{}}}
	idParam := []map[string]any{{"name": "id", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	errContent := subErrContent()
	return map[string]any{
		"/api/v1/subscriptions/user/{userId}": map[string]any{
			"get": getActiveSubscriptionOp(apiKey, errContent),
		},
		"/api/v1/subscriptions/offers": map[string]any{
			"get":  listClubOffersOp(apiKey),
			"post": createClubOfferOp(apiKey, errContent),
		},
		"/api/v1/subscriptions/offers/{id}": map[string]any{
			"get":    getClubOfferOp(idParam, apiKey, errContent),
			"delete": deleteClubOfferOp(idParam, apiKey, errContent),
		},
	}
}

func subErrContent() map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"}}}
}

func subJSON(schema map[string]any) map[string]any {
	return map[string]any{"application/json": map[string]any{"schema": schema}}
}

func subscriptionSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":                map[string]any{"type": "integer"},
		"user_id":           map[string]any{"type": "integer"},
		"subscription_type": map[string]any{"type": "string", "enum": []string{"habbo_club", "builders_club"}},
		"started_at":        map[string]any{"type": "string", "format": "date-time"},
		"duration_days":     map[string]any{"type": "integer"},
		"active":            map[string]any{"type": "boolean"},
		"created_at":        map[string]any{"type": "string", "format": "date-time"},
		"updated_at":        map[string]any{"type": "string", "format": "date-time"},
	}}
}

func clubOfferSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"id":           map[string]any{"type": "integer"},
		"name":         map[string]any{"type": "string"},
		"days":         map[string]any{"type": "integer"},
		"credits":      map[string]any{"type": "integer"},
		"points":       map[string]any{"type": "integer"},
		"points_type":  map[string]any{"type": "integer"},
		"offer_type":   map[string]any{"type": "string", "enum": []string{"HC", "VIP"}},
		"giftable":     map[string]any{"type": "boolean"},
		"enabled":      map[string]any{"type": "boolean"},
	}}
}

func clubOfferRequestSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "days", "offer_type"},
		"properties": map[string]any{
			"name":        map[string]any{"type": "string"},
			"days":        map[string]any{"type": "integer", "minimum": 1},
			"credits":     map[string]any{"type": "integer"},
			"points":      map[string]any{"type": "integer"},
			"points_type": map[string]any{"type": "integer"},
			"offer_type":  map[string]any{"type": "string", "enum": []string{"HC", "VIP"}},
			"giftable":    map[string]any{"type": "boolean"},
			"enabled":     map[string]any{"type": "boolean"},
		}}
}

func getActiveSubscriptionOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Get active subscription", "security": apiKey, "parameters": userParam,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": subJSON(subscriptionSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listClubOffersOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "List club offers", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": subJSON(map[string]any{"type": "array", "items": clubOfferSchema()})}}}
}

func createClubOfferOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Create club offer", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": subJSON(clubOfferRequestSchema())},
		"responses": map[string]any{
			"201": map[string]any{"description": "Created", "content": subJSON(clubOfferSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func getClubOfferOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Get club offer", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": subJSON(clubOfferSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func deleteClubOfferOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Delete club offer", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content"},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}
