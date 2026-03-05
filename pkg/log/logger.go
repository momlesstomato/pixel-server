package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a zap logger built from Config.
func New(cfg Config) (*zap.Logger, error) {
	zapCfg, err := ZapConfig(cfg)
	if err != nil {
		return nil, err
	}
	return zapCfg.Build()
}

// ZapConfig builds zap.Config from Config.
func ZapConfig(cfg Config) (zap.Config, error) {
	if err := cfg.Validate(); err != nil {
		return zap.Config{}, err
	}
	level, _ := zapcore.ParseLevel(cfg.Level)
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.Encoding = cfg.Format
	return zapCfg, nil
}
