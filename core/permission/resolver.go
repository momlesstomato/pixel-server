package permission

import "strings"

// Resolve reports whether one permission is granted by a permission set.
func Resolve(grants map[string]struct{}, permission string) bool {
	if len(grants) == 0 {
		return false
	}
	normalized := strings.TrimSpace(permission)
	if normalized == "" {
		return false
	}
	if _, ok := grants[normalized]; ok {
		return true
	}
	if _, ok := grants[WildcardPermission]; ok {
		return true
	}
	segments := strings.Split(normalized, ".")
	if len(segments) <= 1 {
		return false
	}
	for index := len(segments) - 1; index > 0; index-- {
		candidate := strings.Join(segments[:index], ".") + ".*"
		if _, ok := grants[candidate]; ok {
			return true
		}
	}
	return false
}
