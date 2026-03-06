package cli

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	httpserver "pixelsv/pkg/http"
	"pixelsv/pkg/storage/postgres"
	"pixelsv/pkg/storage/redis"
)

// startupPlan describes role-selected runtime dependencies.
type startupPlan struct {
	// HTTP contains HTTP configuration when HTTP listener is active.
	HTTP *httpserver.Config
	// Postgres contains PostgreSQL configuration when DB is required.
	Postgres *postgres.Config
	// Redis contains Redis configuration when cache/store is required.
	Redis *redis.Config
}

// buildStartupPlan loads package configurations according to active roles.
func buildStartupPlan(v *viper.Viper, roles roleSet) (startupPlan, error) {
	plan := startupPlan{}
	if roles.needsHTTP() {
		if err := httpserver.BindViper(v); err != nil {
			return plan, err
		}
		httpCfg, err := httpserver.FromViper(v)
		if err != nil {
			return plan, err
		}
		plan.HTTP = &httpCfg
	}
	if roles.needsPostgres() {
		if err := postgres.BindViper(v); err != nil {
			return plan, err
		}
		pgCfg, err := postgres.FromViper(v)
		if err != nil {
			return plan, err
		}
		plan.Postgres = &pgCfg
	}
	if roles.needsRedis() {
		if err := redis.BindViper(v); err != nil {
			return plan, err
		}
		rdCfg, err := redis.FromViper(v)
		if err != nil {
			return plan, err
		}
		plan.Redis = &rdCfg
	}
	return plan, nil
}

// runRoleAwareStartup starts dependencies selected by active roles.
func runRoleAwareStartup(ctx context.Context, v *viper.Viper, logger *zap.Logger, roles roleSet) error {
	plan, err := buildStartupPlan(v, roles)
	if err != nil {
		return err
	}
	if plan.Postgres != nil {
		pgSvc, err := postgres.New(ctx, *plan.Postgres)
		if err != nil {
			return err
		}
		defer pgSvc.Close()
	}
	if plan.Redis != nil {
		rdSvc, err := redis.New(*plan.Redis)
		if err != nil {
			return err
		}
		defer rdSvc.Close()
	}
	if plan.HTTP != nil {
		srv, err := httpserver.New(*plan.HTTP, logger)
		if err != nil {
			return err
		}
		return srv.ListenAndServe(ctx)
	}
	<-ctx.Done()
	return nil
}
