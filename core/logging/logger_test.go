package logging

import (
	"bytes"
	"strings"
	"testing"

	"github.com/momlesstomato/pixel-server/core/config"
)

// TestNewBuildsJSONLogger verifies JSON encoder output.
func TestNewBuildsJSONLogger(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := New(config.LoggingConfig{Format: "json", Level: "info"}, buffer)
	if err != nil {
		t.Fatalf("expected logger build to succeed, got error: %v", err)
	}
	logger.Info("json-message")
	output := buffer.String()
	if !strings.Contains(output, "\"msg\":\"json-message\"") {
		t.Fatalf("expected JSON output to contain message field, got: %s", output)
	}
}

// TestNewBuildsConsoleLogger verifies console encoder output.
func TestNewBuildsConsoleLogger(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := New(config.LoggingConfig{Format: "console", Level: "info"}, buffer)
	if err != nil {
		t.Fatalf("expected logger build to succeed, got error: %v", err)
	}
	logger.Info("console-message")
	output := buffer.String()
	if !strings.Contains(output, "console-message") {
		t.Fatalf("expected console output to contain message, got: %s", output)
	}
}

// TestNewRejectsInvalidFormat verifies validation for unsupported formats.
func TestNewRejectsInvalidFormat(t *testing.T) {
	_, err := New(config.LoggingConfig{Format: "xml", Level: "info"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatalf("expected logger build to fail for invalid format")
	}
}

// TestNewRejectsInvalidLevel verifies validation for unsupported levels.
func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := New(config.LoggingConfig{Format: "json", Level: "verbose"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatalf("expected logger build to fail for invalid level")
	}
}

// TestNewUsesStdoutWhenOutputIsNil verifies nil output fallback.
func TestNewUsesStdoutWhenOutputIsNil(t *testing.T) {
	logger, err := New(config.LoggingConfig{Format: "console", Level: "info"}, nil)
	if err != nil {
		t.Fatalf("expected logger build to succeed, got error: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected non-nil logger")
	}
}
