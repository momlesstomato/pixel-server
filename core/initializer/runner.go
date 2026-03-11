package initializer

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/logging"
)

// Runner executes typed startup stages in explicit dependency order.
type Runner struct {
	// config stores configuration startup behavior.
	config config.Stage
	// logger stores logger startup behavior.
	logger logging.Stage
	// http stores HTTP startup behavior.
	http corehttp.Stage
	// websockets stores websocket startup behavior.
	websockets []corehttp.WebSocketStage
}

// NewRunner creates a startup runner with explicit typed stages.
func NewRunner(config config.Stage, logger logging.Stage, http corehttp.Stage, websockets ...corehttp.WebSocketStage) *Runner {
	return &Runner{config: config, logger: logger, http: http, websockets: websockets}
}

// Run executes startup stages and returns initialized runtime dependencies.
func (runner *Runner) Run() (*Runtime, error) {
	if runner.config == nil || runner.logger == nil || runner.http == nil {
		return nil, fmt.Errorf("config, logger and http stages are required")
	}
	loadedConfig, err := runner.config.InitializeConfig()
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.config.Name(), err)
	}
	loadedLogger, err := runner.logger.InitializeLogger(loadedConfig)
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.logger.Name(), err)
	}
	loadedHTTP, err := runner.http.InitializeHTTP(loadedLogger)
	if err != nil {
		return nil, fmt.Errorf("initializer %s failed: %w", runner.http.Name(), err)
	}
	for _, stage := range runner.websockets {
		if err := stage.InitializeWebSocket(loadedHTTP); err != nil {
			return nil, fmt.Errorf("initializer %s failed: %w", stage.Name(), err)
		}
	}
	return &Runtime{Config: loadedConfig, Logger: loadedLogger, HTTP: loadedHTTP}, nil
}
