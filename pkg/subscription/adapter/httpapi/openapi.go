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
		"/api/v1/subscriptions/payday/config": map[string]any{
			"get":   getPaydayConfigOp(apiKey),
			"patch": updatePaydayConfigOp(apiKey, errContent),
		},
		"/api/v1/subscriptions/user/{userId}/payday": map[string]any{
			"get": getPaydayStatusOp(apiKey, errContent),
		},
		"/api/v1/subscriptions/user/{userId}/payday/trigger": map[string]any{
			"post": triggerPaydayOp(apiKey, errContent),
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
		"id":          map[string]any{"type": "integer"},
		"name":        map[string]any{"type": "string"},
		"days":        map[string]any{"type": "integer"},
		"credits":     map[string]any{"type": "integer"},
		"points":      map[string]any{"type": "integer"},
		"points_type": map[string]any{"type": "integer"},
		"offer_type":  map[string]any{"type": "string", "enum": []string{"HC", "VIP"}},
		"giftable":    map[string]any{"type": "boolean"},
		"enabled":     map[string]any{"type": "boolean"},
	}}
}

func paydayConfigSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"interval_days":         map[string]any{"type": "integer"},
		"kickback_percentage":   map[string]any{"type": "number", "format": "double"},
		"flat_credits":          map[string]any{"type": "integer"},
		"minimum_credits_spent": map[string]any{"type": "integer"},
		"streak_bonus_credits":  map[string]any{"type": "integer"},
		"updated_at":            map[string]any{"type": "string", "format": "date-time"},
	}}
}

func paydayStateSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"user_id":                map[string]any{"type": "integer"},
		"first_subscription_at":  map[string]any{"type": "string", "format": "date-time"},
		"next_payday_at":         map[string]any{"type": "string", "format": "date-time"},
		"cycle_credits_spent":    map[string]any{"type": "integer"},
		"reward_streak":          map[string]any{"type": "integer"},
		"total_credits_rewarded": map[string]any{"type": "integer"},
		"total_credits_missed":   map[string]any{"type": "integer"},
		"club_gifts_claimed":     map[string]any{"type": "integer"},
	}}
}

func paydayStatusSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"config":                 paydayConfigSchema(),
		"state":                  paydayStateSchema(),
		"current_hc_streak_days": map[string]any{"type": "integer"},
		"spend_reward_credits":   map[string]any{"type": "integer"},
		"streak_reward_credits":  map[string]any{"type": "integer"},
		"total_reward_credits":   map[string]any{"type": "integer"},
	}}
}

func paydayResultSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"status":         paydayStatusSchema(),
		"reward_credits": map[string]any{"type": "integer"},
		"new_credits":    map[string]any{"type": "integer"},
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

func getPaydayConfigOp(apiKey []map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Get payday config", "security": apiKey,
		"responses": map[string]any{"200": map[string]any{"description": "OK", "content": subJSON(paydayConfigSchema())}}}
}

func updatePaydayConfigOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Update payday config", "security": apiKey,
		"requestBody": map[string]any{"required": true, "content": subJSON(paydayConfigSchema())},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": subJSON(paydayConfigSchema())},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}

func getPaydayStatusOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Get payday status", "security": apiKey, "parameters": userParam,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": subJSON(paydayStatusSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func triggerPaydayOp(apiKey []map[string]any, errContent map[string]any) map[string]any {
	userParam := []map[string]any{{"name": "userId", "in": "path", "required": true, "schema": map[string]any{"type": "integer", "minimum": 1}}}
	requestSchema := map[string]any{"type": "object", "properties": map[string]any{"conn_id": map[string]any{"type": "string"}, "force": map[string]any{"type": "boolean"}}}
	return map[string]any{"tags": []string{"Subscriptions"}, "summary": "Trigger payday", "security": apiKey, "parameters": userParam,
		"requestBody": map[string]any{"required": false, "content": subJSON(requestSchema)},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": subJSON(paydayResultSchema())},
			"404": map[string]any{"description": "Not Found", "content": errContent},
			"409": map[string]any{"description": "Conflict", "content": errContent},
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
