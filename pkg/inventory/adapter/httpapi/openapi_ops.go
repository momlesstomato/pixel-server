package httpapi

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

func addCreditsOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "Add credits", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": invJSON(modifyAmountSchema())},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(map[string]any{"type": "object", "properties": map[string]any{
				"user_id": map[string]any{"type": "integer"},
				"credits": map[string]any{"type": "integer"},
			}})},
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

func addCurrencyOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "Add currency by type", "security": apiKey, "parameters": params,
		"requestBody": map[string]any{"required": true, "content": invJSON(modifyAmountSchema())},
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(currencySchema())},
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

func revokeBadgeOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "Revoke badge", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"204": map[string]any{"description": "No Content", "content": invJSON(map[string]any{"type": "object"})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
			"404": map[string]any{"description": "Not Found", "content": errContent},
		}}
}

func listEffectsOp(params, apiKey []map[string]any, errContent map[string]any) map[string]any {
	return map[string]any{"tags": []string{"Inventory"}, "summary": "List effects", "security": apiKey, "parameters": params,
		"responses": map[string]any{
			"200": map[string]any{"description": "OK", "content": invJSON(map[string]any{"type": "array", "items": effectSchema()})},
			"400": map[string]any{"description": "Bad Request", "content": errContent},
		}}
}
