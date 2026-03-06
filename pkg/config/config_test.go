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
	t.Setenv("PIXELSV_ROLE", "")
	t.Setenv("PIXELSV_INSTANCE_ID", "")
	t.Setenv("NATS_URL", "")
	cfg, err := Load(LoadOptions{EnvFile: "testdata/missing.env"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("expected development env, got %q", cfg.App.Env)
	}
	if cfg.Runtime.Role != "all" {
		t.Fatalf("expected all role, got %q", cfg.Runtime.Role)
	}
	if cfg.Runtime.InstanceID != "pixelsv-local" {
		t.Fatalf("expected default instance id, got %q", cfg.Runtime.InstanceID)
	}
	if cfg.Runtime.NATSURL != "" {
		t.Fatalf("expected empty NATS URL, got %q", cfg.Runtime.NATSURL)
	}
}

// TestLoadFromEnvironment validates env variable override behavior.
func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PIXELSV_ROLE", "gateway,api")
	t.Setenv("PIXELSV_INSTANCE_ID", "gateway-01")
	t.Setenv("NATS_URL", "nats://localhost:4222")
	cfg, err := Load(DefaultLoadOptions())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.App.Env)
	}
	if cfg.Runtime.Role != "gateway,api" {
		t.Fatalf("expected gateway,api role, got %q", cfg.Runtime.Role)
	}
	if cfg.Runtime.InstanceID != "gateway-01" {
		t.Fatalf("expected gateway-01 instance id, got %q", cfg.Runtime.InstanceID)
	}
	if cfg.Runtime.NATSURL != "nats://localhost:4222" {
		t.Fatalf("expected nats url, got %q", cfg.Runtime.NATSURL)
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

// TestRuntimeConfigValidate checks runtime config validation rules.
func TestRuntimeConfigValidate(t *testing.T) {
	if err := (RuntimeConfig{Role: "", InstanceID: "x"}).Validate(); err == nil {
		t.Fatalf("expected empty role validation error")
	}
	if err := (RuntimeConfig{Role: "game", InstanceID: ""}).Validate(); err == nil {
		t.Fatalf("expected empty instance id validation error")
	}
	if err := (RuntimeConfig{Role: "invalid", InstanceID: "x"}).Validate(); err == nil {
		t.Fatalf("expected invalid role validation error")
	}
	if err := (RuntimeConfig{Role: "game,gateway", InstanceID: "x"}).Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// TestLoadFromEnvFile validates env file parsing.
func TestLoadFromEnvFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := "APP_ENV=qa\nPIXELSV_ROLE=api\nPIXELSV_INSTANCE_ID=api-01\nNATS_URL=nats://nats:4222\n"
	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	cfg, err := Load(LoadOptions{EnvFile: envFile})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "qa" {
		t.Fatalf("expected qa env, got %q", cfg.App.Env)
	}
	if cfg.Runtime.Role != "api" || cfg.Runtime.InstanceID != "api-01" {
		t.Fatalf("unexpected runtime config: %+v", cfg.Runtime)
	}
}

// TestFromViperAppliesDefaults checks default filling from tags.
func TestFromViperAppliesDefaults(t *testing.T) {
	v := viper.New()
	v.Set("app.env", "")
	v.Set("runtime.role", "")
	v.Set("runtime.instance_id", "")
	v.Set("runtime.nats_url", "")
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.App.Env != "development" {
		t.Fatalf("unexpected env value: %s", cfg.App.Env)
	}
	if cfg.Runtime.Role != "all" || cfg.Runtime.InstanceID != "pixelsv-local" {
		t.Fatalf("unexpected runtime defaults: %+v", cfg.Runtime)
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
	runtimeTyp := reflect.TypeOf(RuntimeConfig{})
	if runtimeTyp.Field(0).Tag.Get("default") == "" {
		t.Fatalf("expected default tag on RuntimeConfig.Role")
	}
	if runtimeTyp.Field(1).Tag.Get("default") == "" {
		t.Fatalf("expected default tag on RuntimeConfig.InstanceID")
	}
	if _, ok := runtimeTyp.Field(2).Tag.Lookup("default"); !ok {
		t.Fatalf("expected default tag on RuntimeConfig.NATSURL")
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
		Ratio float64 `mapstructure:"ratio" default:"1.5"`
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

// TestParseRoles checks runtime role parsing behavior.
func TestParseRoles(t *testing.T) {
	roles, err := ParseRoles("gateway,api,gateway")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(roles) != 2 || roles[0] != "gateway" || roles[1] != "api" {
		t.Fatalf("unexpected roles: %+v", roles)
	}
	all, err := ParseRoles("all,game")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(all) != 1 || all[0] != "all" {
		t.Fatalf("unexpected all roles result: %+v", all)
	}
	if _, err := ParseRoles("unknown"); err == nil {
		t.Fatalf("expected invalid role error")
	}
	if _, err := ParseRoles("all,unknown"); err == nil {
		t.Fatalf("expected invalid role error")
	}
	if _, err := ParseRoles(" , "); err == nil {
		t.Fatalf("expected empty role error")
	}
}
