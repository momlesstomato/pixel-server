package redis

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// TestFromViper validates env binding and parsing.
func TestFromViper(t *testing.T) {
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("REDIS_KEY_PREFIX", "px")
	t.Setenv("REDIS_SESSION_TTL_SECONDS", "120")
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v.AutomaticEnv()
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.KeyPrefix != "px" || cfg.SessionTTLSeconds != 120 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

// TestConfigValidate checks validation rules.
func TestConfigValidate(t *testing.T) {
	if err := (Config{URL: "", SessionTTLSeconds: 1}).Validate(); err == nil {
		t.Fatalf("expected empty url error")
	}
	if err := (Config{URL: "redis://localhost:6379/0", SessionTTLSeconds: 0}).Validate(); err == nil {
		t.Fatalf("expected ttl error")
	}
}

// TestFromViperDefaults validates default values for optional fields.
func TestFromViperDefaults(t *testing.T) {
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v.Set("storage.redis.url", "redis://localhost:6379/0")
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.KeyPrefix != "pixelsv" || cfg.SessionTTLSeconds != 3600 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}

// TestConfigDefaultTags validates required/optional tag contract.
func TestConfigDefaultTags(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	if typ.Field(0).Tag.Get("default") != "" {
		t.Fatalf("expected URL to be required without default tag")
	}
	if typ.Field(1).Tag.Get("default") == "" || typ.Field(2).Tag.Get("default") == "" {
		t.Fatalf("expected defaults for redis optional fields")
	}
}
