package http

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestNewSetsDefaultReadBufferSize verifies default header buffer behavior.
func TestNewSetsDefaultReadBufferSize(t *testing.T) {
	module := New(Options{})
	if module.App().Config().ReadBufferSize != DefaultReadBufferSize {
		t.Fatalf("expected read buffer size %d, got %d", DefaultReadBufferSize, module.App().Config().ReadBufferSize)
	}
}

// TestNewPreservesCustomReadBufferSize verifies explicit config behavior.
func TestNewPreservesCustomReadBufferSize(t *testing.T) {
	module := New(Options{FiberConfig: fiber.Config{ReadBufferSize: 32 * 1024}})
	if module.App().Config().ReadBufferSize != 32*1024 {
		t.Fatalf("expected read buffer size %d, got %d", 32*1024, module.App().Config().ReadBufferSize)
	}
}
