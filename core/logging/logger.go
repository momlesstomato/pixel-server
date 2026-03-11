package logging

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/momlesstomato/pixel-server/core/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap logger using configuration-driven format and level.
func New(cfg config.LoggingConfig, output io.Writer) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		return nil, fmt.Errorf("parse logging level %q: %w", cfg.Level, err)
	}
	encoder, err := newEncoder(cfg.Format)
	if err != nil {
		return nil, err
	}
	if output == nil {
		output = os.Stdout
	}
	core := zapcore.NewCore(encoder, zapcore.AddSync(output), level)
	return zap.New(core), nil
}

// newEncoder creates an encoder based on the selected output format.
func newEncoder(format string) (zapcore.Encoder, error) {
	switch strings.ToLower(format) {
	case "json":
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), nil
	case "console":
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		return zapcore.NewConsoleEncoder(encoderConfig), nil
	default:
		return nil, fmt.Errorf("unsupported log format %q", format)
	}
}
