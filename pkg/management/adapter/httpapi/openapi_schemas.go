package httpapi

// sessionSchema returns the JSON Schema for one session entry.
func sessionSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"conn_id", "state", "instance_id", "created_at"},
		"properties": map[string]any{
			"conn_id":     map[string]any{"type": "string"},
			"user_id":     map[string]any{"type": "integer"},
			"machine_id":  map[string]any{"type": "string"},
			"state":       map[string]any{"type": "string"},
			"instance_id": map[string]any{"type": "string"},
			"created_at":  map[string]any{"type": "string", "format": "date-time"},
		},
	}
}

// sessionListSchema returns the JSON Schema for a session list response.
func sessionListSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"sessions", "count"},
		"properties": map[string]any{
			"sessions": map[string]any{"type": "array", "items": sessionSchema()},
			"count":    map[string]any{"type": "integer"},
		},
	}
}

// hotelStatusSchema returns the JSON Schema for a hotel status response.
func hotelStatusSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"state", "throw_users"},
		"properties": map[string]any{
			"state":       map[string]any{"type": "string"},
			"close_at":    map[string]any{"type": "string", "format": "date-time"},
			"reopen_at":   map[string]any{"type": "string", "format": "date-time"},
			"throw_users": map[string]any{"type": "boolean"},
		},
	}
}
