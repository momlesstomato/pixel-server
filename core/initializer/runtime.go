package initializer

import (
	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"go.uber.org/zap"
)

// Runtime stores typed outputs produced by startup stages.
type Runtime struct {
	// Config stores loaded application configuration.
	Config *config.Config
	// Logger stores the initialized zap logger.
	Logger *zap.Logger
	// HTTP stores the initialized HTTP module.
	HTTP *corehttp.Module
}
