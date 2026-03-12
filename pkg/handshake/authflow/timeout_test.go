package authflow

import (
	"context"
	"errors"
	"testing"
	"time"
)

// timeoutTransportStub defines close capture behavior for timeout tests.
type timeoutTransportStub struct {
	// closed stores connection identifiers closed by timeout flow.
	closed []string
}

// Send is a no-op for timeout-focused test behavior.
func (stub *timeoutTransportStub) Send(_ string, _ uint16, _ []byte) error {
	return nil
}

// Close captures closed connection identifiers.
func (stub *timeoutTransportStub) Close(connID string, _ int, _ string) error {
	stub.closed = append(stub.closed, connID)
	return nil
}

// TestNewTimeoutUseCaseRejectsNilTransport verifies constructor preconditions.
func TestNewTimeoutUseCaseRejectsNilTransport(t *testing.T) {
	if _, err := NewTimeoutUseCase(nil, time.Second); err == nil {
		t.Fatalf("expected constructor failure for nil transport")
	}
}

// TestTimeoutUseCaseWaitReturnsWhenAuthenticated verifies timeout bypass behavior.
func TestTimeoutUseCaseWaitReturnsWhenAuthenticated(t *testing.T) {
	transport := &timeoutTransportStub{}
	useCase, err := NewTimeoutUseCase(transport, time.Second)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	authenticated := make(chan struct{})
	close(authenticated)
	if waitErr := useCase.Wait(context.Background(), "conn-id", authenticated); waitErr != nil {
		t.Fatalf("expected wait success, got %v", waitErr)
	}
	if len(transport.closed) != 0 {
		t.Fatalf("expected no close call, got %v", transport.closed)
	}
}

// TestTimeoutUseCaseWaitClosesConnectionOnTimeout verifies timeout close behavior.
func TestTimeoutUseCaseWaitClosesConnectionOnTimeout(t *testing.T) {
	transport := &timeoutTransportStub{}
	useCase, err := NewTimeoutUseCase(transport, time.Second)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	trigger := make(chan time.Time, 1)
	useCase.after = func(_ time.Duration) <-chan time.Time { return trigger }
	authenticated := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- useCase.Wait(context.Background(), "conn-id", authenticated) }()
	trigger <- time.Now()
	if waitErr := <-done; !errors.Is(waitErr, ErrAuthTimeoutElapsed) {
		t.Fatalf("expected auth timeout error, got %v", waitErr)
	}
	if len(transport.closed) != 1 || transport.closed[0] != "conn-id" {
		t.Fatalf("expected timeout close call, got %v", transport.closed)
	}
}

// TestTimeoutUseCaseWaitReturnsOnContextCancel verifies context cancellation behavior.
func TestTimeoutUseCaseWaitReturnsOnContextCancel(t *testing.T) {
	transport := &timeoutTransportStub{}
	useCase, err := NewTimeoutUseCase(transport, time.Second)
	if err != nil {
		t.Fatalf("expected constructor success, got %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if waitErr := useCase.Wait(ctx, "conn-id", make(chan struct{})); waitErr != nil {
		t.Fatalf("expected wait success, got %v", waitErr)
	}
	if len(transport.closed) != 0 {
		t.Fatalf("expected no close call, got %v", transport.closed)
	}
}
