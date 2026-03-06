package cli

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
	"pixelsv/pkg/config"
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
	plan, err := buildStartupPlan(v, config.RuntimeConfig{}, roles)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Transport == nil || plan.HTTP == nil || plan.Redis == nil || plan.Postgres != nil {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	defer plan.Transport.Close()
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
	plan, err := buildStartupPlan(v, config.RuntimeConfig{}, roles)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if plan.Transport == nil || plan.HTTP != nil || plan.Postgres == nil || plan.Redis == nil {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	defer plan.Transport.Close()
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
	plan, err := buildStartupPlan(v, config.RuntimeConfig{}, roles)
	if plan.Transport != nil {
		defer plan.Transport.Close()
	}
	if err == nil {
		t.Fatalf("expected api key validation error")
	}
	if !strings.Contains(err.Error(), "api key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestBuildStartupPlanNATSSelection validates distributed transport selection.
func TestBuildStartupPlanNATSSelection(t *testing.T) {
	roles, err := newRoleSet("gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v := viper.New()
	v.Set("http.api_key", "secret")
	v.Set("storage.redis.url", "redis://localhost:6379/0")
	_, err = buildStartupPlan(v, config.RuntimeConfig{NATSURL: "://bad"}, roles)
	if err == nil {
		t.Fatalf("expected nats connection error without server")
	}
}

// TestTransportMode validates transport mode selection behavior.
func TestTransportMode(t *testing.T) {
	all, err := newRoleSet("all")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := transportMode(config.RuntimeConfig{NATSURL: "nats://nats:4222"}, all); got != "local" {
		t.Fatalf("expected local transport mode, got %s", got)
	}
	gateway, err := newRoleSet("gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := transportMode(config.RuntimeConfig{}, gateway); got != "local" {
		t.Fatalf("expected local transport mode, got %s", got)
	}
	if got := transportMode(config.RuntimeConfig{NATSURL: "nats://nats:4222"}, gateway); got != "nats" {
		t.Fatalf("expected nats transport mode, got %s", got)
	}
}
