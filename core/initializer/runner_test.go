package initializer

import (
	"errors"
	"testing"

	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"go.uber.org/zap"
)

// TestRunnerRunExecutesStagesInOrder verifies explicit startup ordering.
func TestRunnerRunExecutesStagesInOrder(t *testing.T) {
	order := make([]string, 0, 4)
	runner := NewRunner(
		testConfigStage{order: &order},
		testLoggerStage{order: &order},
		testHTTPStage{order: &order},
		testWebSocketStage{order: &order},
	)
	if _, err := runner.Run(); err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	expected := []string{"config", "logger", "http", "websocket"}
	for index, item := range expected {
		if len(order) <= index || order[index] != item {
			t.Fatalf("unexpected execution order: %v", order)
		}
	}
}

// TestRunnerRunStopsOnError verifies short-circuit behavior on stage failures.
func TestRunnerRunStopsOnError(t *testing.T) {
	order := make([]string, 0, 4)
	runner := NewRunner(
		testConfigStage{order: &order},
		testLoggerStage{order: &order},
		testHTTPStage{order: &order},
		testWebSocketStage{order: &order, err: errors.New("boom")},
	)
	if _, err := runner.Run(); err == nil {
		t.Fatalf("expected run error")
	}
	if len(order) != 4 {
		t.Fatalf("expected websocket stage execution before error, got %v", order)
	}
}

// TestRunnerRunRequiresCoreStages verifies required stage validation.
func TestRunnerRunRequiresCoreStages(t *testing.T) {
	if _, err := NewRunner(nil, nil, nil).Run(); err == nil {
		t.Fatalf("expected missing stage validation error")
	}
}

// testConfigStage defines a test config startup stage.
type testConfigStage struct {
	// order records execution sequence.
	order *[]string
}

// Name returns the stage name.
func (stage testConfigStage) Name() string { return "config" }

// InitializeConfig records execution and returns a valid config.
func (stage testConfigStage) InitializeConfig() (*config.Config, error) {
	*stage.order = append(*stage.order, "config")
	return &config.Config{Logging: config.LoggingConfig{Format: "json", Level: "info"}}, nil
}

// testLoggerStage defines a test logger startup stage.
type testLoggerStage struct {
	// order records execution sequence.
	order *[]string
}

// Name returns the stage name.
func (stage testLoggerStage) Name() string { return "logger" }

// InitializeLogger records execution and returns a nop logger.
func (stage testLoggerStage) InitializeLogger(_ *config.Config) (*zap.Logger, error) {
	*stage.order = append(*stage.order, "logger")
	return zap.NewNop(), nil
}

// testHTTPStage defines a test HTTP startup stage.
type testHTTPStage struct {
	// order records execution sequence.
	order *[]string
}

// Name returns the stage name.
func (stage testHTTPStage) Name() string { return "http" }

// InitializeHTTP records execution and returns an HTTP module.
func (stage testHTTPStage) InitializeHTTP(logger *zap.Logger) (*corehttp.Module, error) {
	*stage.order = append(*stage.order, "http")
	return corehttp.New(corehttp.Options{Logger: logger}), nil
}

// testWebSocketStage defines a test websocket startup stage.
type testWebSocketStage struct {
	// order records execution sequence.
	order *[]string
	// err returns an optional stage failure.
	err error
}

// Name returns the stage name.
func (stage testWebSocketStage) Name() string { return "websocket" }

// InitializeWebSocket records execution and returns the configured error.
func (stage testWebSocketStage) InitializeWebSocket(_ *corehttp.Module) error {
	*stage.order = append(*stage.order, "websocket")
	return stage.err
}
