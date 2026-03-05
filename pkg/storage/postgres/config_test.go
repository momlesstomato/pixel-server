package postgres

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// TestFromViper validates env binding and parsing.
func TestFromViper(t *testing.T) {
	t.Setenv("POSTGRES_URL", "postgres://user:pass@localhost:5432/pixelsv?sslmode=disable")
	t.Setenv("POSTGRES_MIN_CONNS", "2")
	t.Setenv("POSTGRES_MAX_CONNS", "8")
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v.AutomaticEnv()
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.MinConns != 2 || cfg.MaxConns != 8 {
		t.Fatalf("unexpected pool config: %+v", cfg)
	}
}

// TestConfigValidate checks validation rules.
func TestConfigValidate(t *testing.T) {
	if err := (Config{URL: "", MinConns: 1, MaxConns: 2}).Validate(); err == nil {
		t.Fatalf("expected empty url error")
	}
	if err := (Config{URL: "postgres://x", MinConns: 5, MaxConns: 2}).Validate(); err == nil {
		t.Fatalf("expected invalid bounds error")
	}
	if err := (Config{URL: "postgres://x", MinConns: 1, MaxConns: 2}).Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestConfigDefaultTags validates required/optional tag contract.
func TestConfigDefaultTags(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	if typ.Field(0).Tag.Get("default") != "" {
		t.Fatalf("expected URL to be required without default tag")
	}
	if typ.Field(1).Tag.Get("default") == "" || typ.Field(2).Tag.Get("default") == "" {
		t.Fatalf("expected defaults for pool bounds")
	}
}
