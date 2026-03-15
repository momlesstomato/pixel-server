package domain

import (
	"fmt"
	"regexp"
	"strings"
)

var groupNamePattern = regexp.MustCompile(`^[a-z0-9-]{2,64}$`)
var permissionPattern = regexp.MustCompile(`^[a-z0-9_]+(\.[a-z0-9_]+|\.\*)*$|^\*$`)

// Grant defines one group permission grant.
type Grant struct {
	// GroupID stores owning group identifier.
	GroupID int
	// Permission stores dotted-notation permission string.
	Permission string
}

// ValidateGroupName validates one group name value.
func ValidateGroupName(value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if !groupNamePattern.MatchString(normalized) {
		return "", fmt.Errorf("group name must match [a-z0-9-] and be 2..64 characters")
	}
	return normalized, nil
}

// ValidatePermission validates one dotted-notation permission value.
func ValidatePermission(value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if len(normalized) == 0 || len(normalized) > 128 {
		return "", fmt.Errorf("permission must be 1..128 characters")
	}
	if !permissionPattern.MatchString(normalized) {
		return "", fmt.Errorf("permission must be lowercase dotted notation with optional wildcard suffix")
	}
	return normalized, nil
}
