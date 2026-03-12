package initializer

import (
	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Runtime stores typed outputs produced by startup stages.
type Runtime struct {
	// Config stores loaded application configuration.
	Config *config.Config
	// Redis stores initialized Redis client connectivity.
	Redis *redislib.Client
	// PostgreSQL stores initialized PostgreSQL ORM connectivity.
	PostgreSQL *gorm.DB
	// Logger stores the initialized zap logger.
	Logger *zap.Logger
	// HTTP stores the initialized HTTP module.
	HTTP *corehttp.Module
}
