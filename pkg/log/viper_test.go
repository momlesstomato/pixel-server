package log

import (
	"testing"

	"github.com/spf13/viper"
)

// TestFromViper verifies viper integration defaults and env overrides.
func TestFromViper(t *testing.T) {
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("LOG_LEVEL", "error")
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	v.AutomaticEnv()
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Format != FormatJSON || cfg.Level != "error" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

// TestFromViperDefaults verifies defaults when no environment is set.
func TestFromViperDefaults(t *testing.T) {
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "")
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Format != FormatConsole || cfg.Level != "info" {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
}
