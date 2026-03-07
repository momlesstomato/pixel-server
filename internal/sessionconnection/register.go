package sessionconnection

import (
	"context"

	"go.uber.org/zap"
	transportadapter "pixelsv/internal/sessionconnection/adapters/transport"
	"pixelsv/internal/sessionconnection/app"
	coretransport "pixelsv/pkg/core/transport"
	"pixelsv/pkg/plugin"
)

// Runtime stores session-connection runtime dependencies.
type Runtime struct {
	// Service exposes session-connection application behavior.
	Service *app.Service
}

// Register initializes session-connection adapters and subscriptions.
func Register(ctx context.Context, bus coretransport.Bus, events plugin.EventBus, logger *zap.Logger, cfg Config) (*Runtime, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	service := app.NewService(events, cfg.TelemetryMinInterval)
	subscriber := transportadapter.NewSubscriber(bus, service, logger, transportadapter.Config{
		PingInterval:           cfg.PingInterval,
		PongTimeout:            cfg.PongTimeout,
		AvailabilityOpen:       cfg.AvailabilityOpen,
		AvailabilityOnShutdown: cfg.AvailabilityOnShutdown,
		AvailabilityAuthentic:  cfg.AvailabilityAuthentic,
	})
	if err := subscriber.Start(ctx); err != nil {
		return nil, err
	}
	logger.Info("session-connection realm registered")
	return &Runtime{Service: service}, nil
}
