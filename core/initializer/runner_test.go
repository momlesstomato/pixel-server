package initializer

import (
	"errors"
	"testing"

	"github.com/momlesstomato/pixel-server/core/app"
	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/logging"
	"github.com/momlesstomato/pixel-server/core/postgres"
	"github.com/momlesstomato/pixel-server/core/redis"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRunnerRunExecutesStagesInOrder verifies explicit startup ordering.
func TestRunnerRunExecutesStagesInOrder(t *testing.T) {
	order := make([]string, 0, 6)
	runner := NewRunner(testConfigStage{&order}, testRedisStage{&order}, testLoggerStage{&order}, testPostgreSQLStage{&order}, testHTTPStage{&order}, testWebSocketStage{order: &order})
	runtime, err := runner.Run()
	if err != nil {
		t.Fatalf("expected run success, got %v", err)
	}
	_ = runtime.Redis.Close()
	sqlDatabase, _ := runtime.PostgreSQL.DB()
	_ = sqlDatabase.Close()
	expected := []string{"config", "redis", "logger", "postgres", "http", "websocket"}
	for index, item := range expected {
		if len(order) <= index || order[index] != item {
			t.Fatalf("unexpected execution order: %v", order)
		}
	}
}

// TestRunnerRunStopsOnError verifies short-circuit behavior on stage failures.
func TestRunnerRunStopsOnError(t *testing.T) {
	order := make([]string, 0, 6)
	runner := NewRunner(testConfigStage{&order}, testRedisStage{&order}, testLoggerStage{&order}, testPostgreSQLStage{&order}, testHTTPStage{&order}, testWebSocketStage{order: &order, err: errors.New("boom")})
	if _, err := runner.Run(); err == nil {
		t.Fatalf("expected run error")
	}
	if len(order) != 6 {
		t.Fatalf("expected websocket stage execution before error, got %v", order)
	}
}

// TestRunnerRunRequiresCoreStages verifies required stage validation.
func TestRunnerRunRequiresCoreStages(t *testing.T) {
	if _, err := NewRunner(nil, nil, nil, nil, nil).Run(); err == nil {
		t.Fatalf("expected missing stage validation error")
	}
}

// testConfigStage defines a test config startup stage.
type testConfigStage struct{ order *[]string }

// Name returns the stage name.
func (stage testConfigStage) Name() string { return "config" }

// InitializeConfig records execution and returns a valid config.
func (stage testConfigStage) InitializeConfig() (*config.Config, error) {
	*stage.order = append(*stage.order, "config")
	return &config.Config{App: config.AppConfig{APIKey: "test-key"}, Logging: config.LoggingConfig{Format: "json", Level: "info"}, PostgreSQL: config.PostgreSQLConfig{DSN: "postgres://postgres:postgres@127.0.0.1:5432/pixel_server?sslmode=disable"}}, nil
}

// testLoggerStage defines a test logger startup stage.
type testLoggerStage struct{ order *[]string }

// Name returns the stage name.
func (stage testLoggerStage) Name() string { return "logger" }

// InitializeLogger records execution and returns a nop logger.
func (stage testLoggerStage) InitializeLogger(_ logging.Config) (*zap.Logger, error) {
	*stage.order = append(*stage.order, "logger")
	return zap.NewNop(), nil
}

// testRedisStage defines a test redis startup stage.
type testRedisStage struct{ order *[]string }

// Name returns the stage name.
func (stage testRedisStage) Name() string { return "redis" }

// InitializeRedis records execution and returns a redis client.
func (stage testRedisStage) InitializeRedis(_ redis.Config) (*redislib.Client, error) {
	*stage.order = append(*stage.order, "redis")
	return redislib.NewClient(&redislib.Options{Addr: "localhost:6379"}), nil
}

// testPostgreSQLStage defines a test PostgreSQL startup stage.
type testPostgreSQLStage struct{ order *[]string }

// Name returns the stage name.
func (stage testPostgreSQLStage) Name() string { return "postgres" }

// InitializePostgreSQL records execution and returns sqlite-backed orm connectivity.
func (stage testPostgreSQLStage) InitializePostgreSQL(_ postgres.Config) (*gorm.DB, error) {
	*stage.order = append(*stage.order, "postgres")
	return gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
}

// testHTTPStage defines a test HTTP startup stage.
type testHTTPStage struct{ order *[]string }

// Name returns the stage name.
func (stage testHTTPStage) Name() string { return "http" }

// InitializeHTTP records execution and returns an HTTP module.
func (stage testHTTPStage) InitializeHTTP(_ app.Config, logger *zap.Logger) (*corehttp.Module, error) {
	*stage.order = append(*stage.order, "http")
	return corehttp.New(corehttp.Options{Logger: logger}), nil
}

// testWebSocketStage defines a test websocket startup stage.
type testWebSocketStage struct {
	order *[]string
	err   error
}

// Name returns the stage name.
func (stage testWebSocketStage) Name() string { return "websocket" }

// InitializeWebSocket records execution and returns the configured error.
func (stage testWebSocketStage) InitializeWebSocket(_ *corehttp.Module) error {
	*stage.order = append(*stage.order, "websocket")
	return stage.err
}
