package command

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewModerationCommand verifies command tree creation.
func TestNewModerationCommand(t *testing.T) {
	cmd := NewModerationCommand(Dependencies{Output: &bytes.Buffer{}})
	assert.Equal(t, "moderation", cmd.Use)
	assert.Len(t, cmd.Commands(), 7)
}

// TestNewModerationCommandSubcommands verifies subcommand names.
func TestNewModerationCommandSubcommands(t *testing.T) {
	cmd := NewModerationCommand(Dependencies{Output: &bytes.Buffer{}})
	names := make([]string, len(cmd.Commands()))
	for i, c := range cmd.Commands() {
		names[i] = c.Use
	}
	assert.Contains(t, names, "list")
	assert.Contains(t, names, "ban")
	assert.Contains(t, names, "unban")
	assert.Contains(t, names, "history")
	assert.Contains(t, names, "alerts")
}
