package command

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestFormatLogLine verifies chat log line formatting.
func TestFormatLogLine(t *testing.T) {
	at := time.Date(2025, 6, 15, 14, 30, 5, 0, time.UTC)
	line := formatLogLine(at, "talk", "alice", "hello world")
	assert.Equal(t, "[14:30:05] [TALK] alice: hello world", line)
}

// TestFormatLogLine_Shout verifies shout type is uppercased.
func TestFormatLogLine_Shout(t *testing.T) {
	at := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	line := formatLogLine(at, "shout", "bob", "hi")
	assert.Equal(t, "[00:00:00] [SHOUT] bob: hi", line)
}

// TestParseDateWindow_Empty verifies default date window is today.
func TestParseDateWindow_Empty(t *testing.T) {
	from, to, err := parseDateWindow("")
	assert.NoError(t, err)
	assert.Equal(t, 0, from.Hour())
	assert.Equal(t, 0, from.Minute())
	assert.True(t, to.After(from))
}

// TestParseDateWindow_SpecificDate verifies specific date parsing.
func TestParseDateWindow_SpecificDate(t *testing.T) {
	from, to, err := parseDateWindow("2025-03-15")
	assert.NoError(t, err)
	assert.Equal(t, 2025, from.Year())
	assert.Equal(t, time.March, from.Month())
	assert.Equal(t, 15, from.Day())
	assert.True(t, to.After(from))
}

// TestParseDateWindow_InvalidFormat verifies error on bad date.
func TestParseDateWindow_InvalidFormat(t *testing.T) {
	_, _, err := parseDateWindow("not-a-date")
	assert.Error(t, err)
}

// TestStartOfDay verifies time truncation.
func TestStartOfDay(t *testing.T) {
	input := time.Date(2025, 6, 15, 14, 30, 5, 999, time.UTC)
	result := startOfDay(input)
	assert.Equal(t, 0, result.Hour())
	assert.Equal(t, 0, result.Minute())
	assert.Equal(t, 0, result.Second())
	assert.Equal(t, 15, result.Day())
}

// TestNewRoomCommand verifies command tree creation.
func TestNewRoomCommand(t *testing.T) {
	cmd := NewRoomCommand(Dependencies{Output: &bytes.Buffer{}})
	assert.Equal(t, "room", cmd.Use)
	assert.True(t, len(cmd.Commands()) > 0)
}

// TestParsePositiveInt_Valid verifies positive integer parsing.
func TestParsePositiveInt_Valid(t *testing.T) {
	id, err := parsePositiveInt("42")
	assert.NoError(t, err)
	assert.Equal(t, 42, id)
}

// TestParsePositiveInt_Zero verifies zero is rejected.
func TestParsePositiveInt_Zero(t *testing.T) {
	_, err := parsePositiveInt("0")
	assert.Error(t, err)
}

// TestParsePositiveInt_Negative verifies negative is rejected.
func TestParsePositiveInt_Negative(t *testing.T) {
	_, err := parsePositiveInt("-5")
	assert.Error(t, err)
}
