package httpapi

// groupDetailsSchema returns the JSON Schema for a permission group with permissions.
func groupDetailsSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"Group": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"ID":            map[string]any{"type": "integer"},
					"Name":          map[string]any{"type": "string"},
					"DisplayName":   map[string]any{"type": "string"},
					"Priority":      map[string]any{"type": "integer"},
					"ClubLevel":     map[string]any{"type": "integer"},
					"SecurityLevel": map[string]any{"type": "integer"},
					"IsAmbassador":  map[string]any{"type": "boolean"},
					"IsDefault":     map[string]any{"type": "boolean"},
				},
			},
			"Permissions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		},
	}
}

// groupListSchema returns the JSON Schema for a group list response.
func groupListSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"groups": map[string]any{"type": "array", "items": groupDetailsSchema()}, "count": map[string]any{"type": "integer"}},
	}
}

// permissionListSchema returns the JSON Schema for a permission list response.
func permissionListSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{"permissions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "count": map[string]any{"type": "integer"}},
	}
}

// userAccessSchema returns the JSON Schema for a user access assignment response.
func userAccessSchema() map[string]any {
	groupProps := map[string]any{
		"ID":            map[string]any{"type": "integer"},
		"Name":          map[string]any{"type": "string"},
		"DisplayName":   map[string]any{"type": "string"},
		"Priority":      map[string]any{"type": "integer"},
		"ClubLevel":     map[string]any{"type": "integer"},
		"SecurityLevel": map[string]any{"type": "integer"},
		"IsAmbassador":  map[string]any{"type": "boolean"},
		"IsDefault":     map[string]any{"type": "boolean"},
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"UserID":       map[string]any{"type": "integer"},
			"PrimaryGroup": map[string]any{"type": "object", "properties": groupProps},
			"GroupIDs":     map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
		},
	}
}

// createGroupBody returns the requestBody schema for group creation.
func createGroupBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"name"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "display_name": map[string]any{"type": "string"}, "priority": map[string]any{"type": "integer"}, "club_level": map[string]any{"type": "integer"}, "security_level": map[string]any{"type": "integer"}, "is_ambassador": map[string]any{"type": "boolean"}, "is_default": map[string]any{"type": "boolean"}}}}}}
}

// patchGroupBody returns the requestBody schema for group update.
func patchGroupBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"name": map[string]any{"type": "string"}, "display_name": map[string]any{"type": "string"}, "priority": map[string]any{"type": "integer"}}}}}}
}

// addPermissionsBody returns the requestBody schema for adding permissions.
func addPermissionsBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"permissions"}, "properties": map[string]any{"permissions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}}}}
}

// singleGroupBody returns the requestBody schema for single group assignment.
func singleGroupBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"group_id"}, "properties": map[string]any{"group_id": map[string]any{"type": "integer"}}}}}}
}

// multiGroupBody returns the requestBody schema for multi-group assignment.
func multiGroupBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"group_ids"}, "properties": map[string]any{"group_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}}}}}}}
}
