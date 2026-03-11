package initializer

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/logging"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
)

// Runner executes typed startup stages in explicit dependency order.
type Runner struct {
	// config stores configuration startup behavior.
	config config.Stage
	// logger stores logger startup behavior.
	logger logging.Stage
	// redis stores redis startup behavior.
	redis rediscore.Stage
	// http stores HTTP startup behavior.
	http corehttp.Stage
	// websockets stores websocket startup behavior.
	websockets []corehttp.WebSocketStage
}

// NewRunner creates a startup runner with explicit typed stages.
func NewRunner(config config.Stage, redis rediscore.Stage, logger logging.Stage, http corehttp.Stage, websockets ...corehttp.WebSocketStage) *Runner {
	return &Runner{config: config, redis: redis, logger: logger, http: http, websockets: websockets}
}

// Run executes startup stages and returns initialized runtime dependencies.
func (runner *Runner) Run() (*Runtime, error) {
	if runner.config == nil || runner.redis == nil || runner.logger == nil || runner.http == nil {
		return nil, fmt.Errorf("config, redis, logger and http stages are required")
	}
	loadedConfig, err := runner.config.InitializeConfig()
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.config.Name(), err)
	}
	loadedRedis, err := runner.redis.InitializeRedis(loadedConfig.Redis)
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.redis.Name(), err)
	}
	loadedLogger, err := runner.logger.InitializeLogger(loadedConfig.Logging)
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.logger.Name(), err)
	}
	loadedHTTP, err := runner.http.InitializeHTTP(loadedConfig.App, loadedLogger)
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.http.Name(), err)
	}
	for _, stage := range runner.websockets {
		if err := stage.InitializeWebSocket(loadedHTTP); err != nil {
			return nil, fmt.Errorf("initializer %s failed: %w", stage.Name(), err)
		}
	}
	return &Runtime{Config: loadedConfig, Redis: loadedRedis, Logger: loadedLogger, HTTP: loadedHTTP}, nil
}
