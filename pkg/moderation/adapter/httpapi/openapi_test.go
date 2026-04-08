package httpapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOpenAPIPathsReturnsExpectedRoutes verifies OpenAPI path keys.
func TestOpenAPIPathsReturnsExpectedRoutes(t *testing.T) {
	paths := OpenAPIPaths()
	assert.Contains(t, paths, "/api/v1/moderation/actions")
	assert.Contains(t, paths, "/api/v1/moderation/alerts")
	assert.Contains(t, paths, "/api/v1/moderation/actions/{id}")
	assert.Contains(t, paths, "/api/v1/moderation/users/{userId}/actions")
	assert.Contains(t, paths, "/api/v1/moderation/actions/{id}/deactivate")
}

// TestRegisterRoutesNilModule verifies nil module returns error.
func TestRegisterRoutesNilModule(t *testing.T) {
	err := RegisterRoutes(nil, nil)
	assert.Error(t, err)
}
