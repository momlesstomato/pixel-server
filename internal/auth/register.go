package auth

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	httpadapter "pixelsv/internal/auth/adapters/http"
	"pixelsv/internal/auth/adapters/memory"
	transportadapter "pixelsv/internal/auth/adapters/transport"
	"pixelsv/internal/auth/app"
	coretransport "pixelsv/pkg/core/transport"
)

// Runtime holds auth realm runtime dependencies.
type Runtime struct {
	// Service exposes auth application behavior.
	Service *app.Service
}

// Register initializes auth realm adapters and subscriptions.
func Register(ctx context.Context, fiberApp *fiber.App, bus coretransport.Bus, logger *zap.Logger, apiKey string) (*Runtime, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	service := app.NewService(memory.NewTicketStore())
	if fiberApp != nil && apiKey != "" {
		httpadapter.RegisterRoutes(fiberApp, service, apiKey)
	}
	subscriber := transportadapter.NewSubscriber(bus, service, logger)
	if err := subscriber.Start(ctx); err != nil {
		return nil, err
	}
	logger.Info("auth realm registered")
	return &Runtime{Service: service}, nil
}
