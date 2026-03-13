package plugin

import (
	sdk "github.com/momlesstomato/pixel-sdk"
	"go.uber.org/zap"
)

// NewServerImplForTest creates a Server implementation for testing purposes.
func NewServerImplForTest(name string, dispatcher *Dispatcher, deps ServerDependencies, logger *zap.Logger) sdk.Server {
	return newServerImpl(name, dispatcher, deps, logger)
}
