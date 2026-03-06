package cli

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestBuildStartupPlanGatewaySkipsPostgres checks gateway-only dependency selection.
func TestBuildStartupPlanGatewaySkipsPostgres(t *testing.T) {
	roles, err := newRoleSet("gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v := viper.New()
	v.Set("http.api_key", "secret")
	v.Set("storage.redis.url", "redis://localhost:6379/0")
	plan, err := buildStartupPlan(v, roles)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.HTTP == nil || plan.Redis == nil || plan.Postgres != nil {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

// TestBuildStartupPlanGameSkipsHTTP checks game-only startup without API key requirement.
func TestBuildStartupPlanGameSkipsHTTP(t *testing.T) {
	roles, err := newRoleSet("game")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v := viper.New()
	v.Set("storage.postgres.url", "postgres://u:p@localhost:5432/pixelsv?sslmode=disable")
	v.Set("storage.redis.url", "redis://localhost:6379/0")
	plan, err := buildStartupPlan(v, roles)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.HTTP != nil || plan.Postgres == nil || plan.Redis == nil {
		t.Fatalf("unexpected plan: %+v", plan)
	}
}

// TestBuildStartupPlanAPIRequiresAPIKey validates API key requirement when HTTP is active.
func TestBuildStartupPlanAPIRequiresAPIKey(t *testing.T) {
	roles, err := newRoleSet("api")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v := viper.New()
	v.Set("storage.postgres.url", "postgres://u:p@localhost:5432/pixelsv?sslmode=disable")
	v.Set("storage.redis.url", "redis://localhost:6379/0")
	_, err = buildStartupPlan(v, roles)
	if err == nil {
		t.Fatalf("expected api key validation error")
	}
	if !strings.Contains(err.Error(), "api key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
