package main

import (
	"bytes"
	"testing"
)

// TestRunReturnsSuccessForHelp verifies successful root command execution.
func TestRunReturnsSuccessForHelp(t *testing.T) {
	if code := run([]string{"--help"}, bytes.NewBuffer(nil), bytes.NewBuffer(nil)); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

// TestRunReturnsFailureForUnknownFlag verifies command failure exit behavior.
func TestRunReturnsFailureForUnknownFlag(t *testing.T) {
	if code := run([]string{"--unknown"}, bytes.NewBuffer(nil), bytes.NewBuffer(nil)); code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
