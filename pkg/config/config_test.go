package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// TestLoadDefaults validates default configuration values.
func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	cfg, err := Load(LoadOptions{EnvFile: "testdata/missing.env"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("expected development env, got %q", cfg.App.Env)
	}
}

// TestLoadFromEnvironment validates env variable override behavior.
func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	cfg, err := Load(DefaultLoadOptions())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.App.Env)
	}
}

// TestAppConfigValidate checks app config validation rules.
func TestAppConfigValidate(t *testing.T) {
	if err := (AppConfig{Env: ""}).Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
	if err := (AppConfig{Env: "test"}).Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestLoadFromEnvFile validates env file parsing.
func TestLoadFromEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envFile, []byte("APP_ENV=qa\n"), 0o644); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	cfg, err := Load(LoadOptions{EnvFile: envFile})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "qa" {
		t.Fatalf("expected qa env, got %q", cfg.App.Env)
	}
}

// TestFromViperAppliesDefaults checks default filling from tags.
func TestFromViperAppliesDefaults(t *testing.T) {
	v := viper.New()
	v.Set("app.env", "")
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("unexpected env value: %s", cfg.App.Env)
	}
}

// TestApplyDefaultsFromTags validates default tag application.
func TestApplyDefaultsFromTags(t *testing.T) {
	type nested struct {
		Port int `mapstructure:"port" default:"8080"`
	}
	type sample struct {
		App  nested `mapstructure:"app"`
		Name string `mapstructure:"name" default:"pixelsv"`
	}
	v := viper.New()
	if err := ApplyDefaultsFromTags(v, "x", sample{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.GetInt("x.app.port") != 8080 || v.GetString("x.name") != "pixelsv" {
		t.Fatalf("unexpected defaults")
	}
}

// TestApplyDefaultsFromTagsInvalidSample checks non-struct rejection.
func TestApplyDefaultsFromTagsInvalidSample(t *testing.T) {
	v := viper.New()
	if err := ApplyDefaultsFromTags(v, "", "nope"); err == nil {
		t.Fatalf("expected error")
	}
}

// TestFillDefaultsFromTags validates struct default filling behavior.
func TestFillDefaultsFromTags(t *testing.T) {
	cfg := AppConfig{}
	if err := FillDefaultsFromTags(&cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Env != "development" {
		t.Fatalf("unexpected default env: %s", cfg.Env)
	}
}

// TestConfigDefaultTags validates default-tag contract for shared config.
func TestConfigDefaultTags(t *testing.T) {
	typ := reflect.TypeOf(AppConfig{})
	if typ.Field(0).Tag.Get("default") == "" {
		t.Fatalf("expected default tag on AppConfig.Env")
	}
}

// TestDefaultLoadOptions checks default load options.
func TestDefaultLoadOptions(t *testing.T) {
	opts := DefaultLoadOptions()
	if opts.EnvFile != ".env" {
		t.Fatalf("unexpected env file: %s", opts.EnvFile)
	}
}

// TestLoadInvalidEnvFile checks env file read error path.
func TestLoadInvalidEnvFile(t *testing.T) {
	if _, err := Load(LoadOptions{EnvFile: t.TempDir()}); err == nil {
		t.Fatalf("expected env file read error")
	}
}

// TestFillDefaultsFromTagsInvalidTarget checks invalid target handling.
func TestFillDefaultsFromTagsInvalidTarget(t *testing.T) {
	if err := FillDefaultsFromTags((*AppConfig)(nil)); err == nil {
		t.Fatalf("expected nil pointer error")
	}
}

// TestApplyDefaultsFromTagsUnsupportedType checks unsupported kind handling.
func TestApplyDefaultsFromTagsUnsupportedType(t *testing.T) {
	type bad struct {
		Enabled bool `mapstructure:"enabled" default:"true"`
	}
	v := viper.New()
	if err := ApplyDefaultsFromTags(v, "", bad{}); err == nil {
		t.Fatalf("expected unsupported kind error")
	}
}

// TestApplyDefaultsFromTagsAdditionalPaths covers int32 and inferred key paths.
func TestApplyDefaultsFromTagsAdditionalPaths(t *testing.T) {
	type sample struct {
		Count int32  `mapstructure:"count" default:"7"`
		Name  string `default:"pixelsv"`
	}
	v := viper.New()
	if err := ApplyDefaultsFromTags(v, "", sample{}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if v.GetInt("count") != 7 || v.GetString("name") != "pixelsv" {
		t.Fatalf("unexpected defaults from additional paths")
	}
}

// TestFillDefaultsFromTagsNonPointer checks non-pointer target handling.
func TestFillDefaultsFromTagsNonPointer(t *testing.T) {
	cfg := AppConfig{}
	if err := FillDefaultsFromTags(cfg); err == nil {
		t.Fatalf("expected non-pointer error")
	}
}
