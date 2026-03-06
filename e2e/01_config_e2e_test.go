package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"pixelsv/pkg/config"
	logpkg "pixelsv/pkg/log"
	"pixelsv/pkg/storage/postgres"
	"pixelsv/pkg/storage/redis"
)

// Test01ConfigurationComposition validates config composition flow.
func Test01ConfigurationComposition(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := "" +
		"APP_ENV=staging\n" +
		"PIXELSV_ROLE=gateway,api\n" +
		"PIXELSV_INSTANCE_ID=gw-01\n" +
		"NATS_URL=nats://nats:4222\n" +
		"LOG_FORMAT=json\n" +
		"LOG_LEVEL=warn\n" +
		"POSTGRES_URL=postgres://u:p@localhost:5432/pixelsv?sslmode=disable\n" +
		"POSTGRES_MIN_CONNS=1\n" +
		"POSTGRES_MAX_CONNS=5\n" +
		"REDIS_URL=redis://localhost:6379/0\n" +
		"REDIS_KEY_PREFIX=px\n" +
		"REDIS_SESSION_TTL_SECONDS=90\n"
	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v, err := config.NewViper(config.LoadOptions{EnvFile: envFile})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := logpkg.BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := postgres.BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := redis.BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	baseCfg, err := config.FromViper(v)
	if err != nil || baseCfg.App.Env != "staging" {
		t.Fatalf("unexpected base config: %+v %v", baseCfg, err)
	}
	if baseCfg.Runtime.Role != "gateway,api" || baseCfg.Runtime.InstanceID != "gw-01" {
		t.Fatalf("unexpected runtime config: %+v", baseCfg.Runtime)
	}
	logCfg, err := logpkg.FromViper(v)
	if err != nil || logCfg.Format != logpkg.FormatJSON {
		t.Fatalf("unexpected log config: %+v %v", logCfg, err)
	}
	if _, err := logpkg.New(logCfg); err != nil {
		t.Fatalf("expected logger build success, got %v", err)
	}
	pgCfg, err := postgres.FromViper(v)
	if err != nil || pgCfg.MaxConns != 5 {
		t.Fatalf("unexpected postgres config: %+v %v", pgCfg, err)
	}
	rdCfg, err := redis.FromViper(v)
	if err != nil || rdCfg.KeyPrefix != "px" {
		t.Fatalf("unexpected redis config: %+v %v", rdCfg, err)
	}
}
