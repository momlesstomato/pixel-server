package cli

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"pixelsv/pkg/config"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/factory"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/storage/postgres"
	"pixelsv/pkg/storage/redis"
)

// startupPlan describes role-selected runtime dependencies.
type startupPlan struct {
	// Transport is the runtime inter-module transport bus.
	Transport transport.Bus
	// HTTP contains HTTP configuration when HTTP listener is active.
	HTTP *httpserver.Config
	// Postgres contains PostgreSQL configuration when DB is required.
	Postgres *postgres.Config
	// Redis contains Redis configuration when cache/store is required.
	Redis *redis.Config
}

// buildStartupPlan loads package configurations according to active roles.
func buildStartupPlan(v *viper.Viper, runtimeCfg config.RuntimeConfig, roles roleSet) (startupPlan, error) {
	plan := startupPlan{}
	bus, err := factory.New(factory.Config{
		NATSURL:    runtimeCfg.NATSURL,
		ForceLocal: roles.forceLocalTransport(),
	})
	if err != nil {
		return plan, err
	}
	plan.Transport = bus
	fail := func(err error) (startupPlan, error) {
		_ = plan.Transport.Close()
		return startupPlan{}, err
	}
	if roles.needsHTTP() {
		if err := httpserver.BindViper(v); err != nil {
			return fail(err)
		}
		httpCfg, err := httpserver.FromViper(v)
		if err != nil {
			return fail(err)
		}
		plan.HTTP = &httpCfg
	}
	if roles.needsPostgres() {
		if err := postgres.BindViper(v); err != nil {
			return fail(err)
		}
		pgCfg, err := postgres.FromViper(v)
		if err != nil {
			return fail(err)
		}
		plan.Postgres = &pgCfg
	}
	if roles.needsRedis() {
		if err := redis.BindViper(v); err != nil {
			return fail(err)
		}
		rdCfg, err := redis.FromViper(v)
		if err != nil {
			return fail(err)
		}
		plan.Redis = &rdCfg
	}
	return plan, nil
}

// runRoleAwareStartup starts dependencies selected by active roles.
func runRoleAwareStartup(ctx context.Context, v *viper.Viper, runtimeCfg config.RuntimeConfig, logger *zap.Logger, roles roleSet) error {
	plan, err := buildStartupPlan(v, runtimeCfg, roles)
	if err != nil {
		return err
	}
	defer plan.Transport.Close()
	logger.Info(
		"runtime startup initialized",
		zap.Strings("roles", roles.names()),
		zap.String("transport_mode", transportMode(runtimeCfg, roles)),
	)
	if plan.Postgres != nil {
		pgSvc, err := postgres.New(ctx, *plan.Postgres)
		if err != nil {
			return err
		}
		defer pgSvc.Close()
		logger.Info("postgres service started")
	}
	if plan.Redis != nil {
		rdSvc, err := redis.New(*plan.Redis)
		if err != nil {
			return err
		}
		defer rdSvc.Close()
		logger.Info("redis service started")
	}
	if plan.HTTP != nil {
		logger.Info("http service started", zap.String("address", plan.HTTP.Address))
		srv, err := httpserver.New(*plan.HTTP, logger)
		if err != nil {
			return err
		}
		return srv.ListenAndServe(ctx)
	}
	logger.Info("runtime started without http listener")
	<-ctx.Done()
	logger.Info("runtime shutdown requested")
	return nil
}

// transportMode reports whether runtime transport is local or nats-backed.
func transportMode(runtimeCfg config.RuntimeConfig, roles roleSet) string {
	if roles.forceLocalTransport() || runtimeCfg.NATSURL == "" {
		return "local"
	}
	return "nats"
}
