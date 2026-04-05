package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestActionScopeValues verifies scope constant values.
func TestActionScopeValues(t *testing.T) {
	assert.Equal(t, ActionScope("room"), ScopeRoom)
	assert.Equal(t, ActionScope("hotel"), ScopeHotel)
}

// TestActionTypeValues verifies action type constant values.
func TestActionTypeValues(t *testing.T) {
	assert.Equal(t, ActionType("kick"), TypeKick)
	assert.Equal(t, ActionType("ban"), TypeBan)
	assert.Equal(t, ActionType("mute"), TypeMute)
	assert.Equal(t, ActionType("warn"), TypeWarn)
}

// TestErrorsNotNil verifies domain errors are defined.
func TestErrorsNotNil(t *testing.T) {
	assert.NotNil(t, ErrActionNotFound)
	assert.NotNil(t, ErrCannotDeleteHotelAction)
	assert.NotNil(t, ErrAlreadyInactive)
	assert.NotNil(t, ErrInvalidScope)
	assert.NotNil(t, ErrMissingTarget)
}
