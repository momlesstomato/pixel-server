package logging

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewBuildsJSONLogger verifies JSON encoder output.
func TestNewBuildsJSONLogger(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	logger, err := New(Config{Format: "json", Level: "info"}, buffer)
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
	logger, err := New(Config{Format: "console", Level: "info"}, buffer)
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
	_, err := New(Config{Format: "xml", Level: "info"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatalf("expected logger build to fail for invalid format")
	}
}

// TestNewRejectsInvalidLevel verifies validation for unsupported levels.
func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := New(Config{Format: "json", Level: "verbose"}, bytes.NewBuffer(nil))
	if err == nil {
		t.Fatalf("expected logger build to fail for invalid level")
	}
}

// TestNewUsesStdoutWhenOutputIsNil verifies nil output fallback.
func TestNewUsesStdoutWhenOutputIsNil(t *testing.T) {
	logger, err := New(Config{Format: "console", Level: "info"}, nil)
	if err != nil {
		t.Fatalf("expected logger build to succeed, got error: %v", err)
	}
	if logger == nil {
		t.Fatalf("expected non-nil logger")
	}
}

// TestInitializerBuildsLogger verifies package-owned initializer behavior.
func TestInitializerBuildsLogger(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	loaded := Config{Format: "json", Level: "info"}
	logger, err := (Initializer{Output: buffer}).InitializeLogger(loaded)
	if err != nil {
		t.Fatalf("expected initializer success, got %v", err)
	}
	logger.Info("hello")
	if !strings.Contains(buffer.String(), "\"msg\":\"hello\"") {
		t.Fatalf("expected logger output to contain message, got %s", buffer.String())
	}
}

// TestInitializerRejectsEmptyConfig verifies config precondition checks.
func TestInitializerRejectsEmptyConfig(t *testing.T) {
	if _, err := (Initializer{}).InitializeLogger(Config{}); err == nil {
		t.Fatalf("expected initializer error for empty logging config")
	}
}
