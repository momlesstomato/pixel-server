package connection

import (
	"errors"
	"testing"
)

// TestDisposeFuncExecutesWrappedFunction verifies function adapter execution.
func TestDisposeFuncExecutesWrappedFunction(t *testing.T) {
	called := false
	disposable := DisposeFunc(func() error {
		called = true
		return nil
	})
	if err := disposable.Dispose(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatalf("expected wrapped function to be called")
	}
}

// TestDisposeFuncReturnsWrappedError verifies error propagation.
func TestDisposeFuncReturnsWrappedError(t *testing.T) {
	expected := errors.New("dispose failure")
	disposable := DisposeFunc(func() error {
		return expected
	})
	if err := disposable.Dispose(); !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
