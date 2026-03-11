package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// TestApplyAliasValuesMapsUnderscoreKeys verifies .env key alias mapping into nested keys.
func TestApplyAliasValuesMapsUnderscoreKeys(t *testing.T) {
	instance := viper.New()
	instance.Set("APP_PORT", "8123")
	applyAliasValues(instance, []string{"app.port"}, "")
	if got := instance.GetInt("app.port"); got != 8123 {
		t.Fatalf("expected alias to map APP_PORT to app.port, got %d", got)
	}
}

// TestMissingRequiredKeysHandlesExplicitFalse verifies false bool env values are treated as set.
func TestMissingRequiredKeysHandlesExplicitFalse(t *testing.T) {
	t.Setenv("CFG_BOOL_FEATURE_ENABLED", "false")
	instance := viper.New()
	instance.SetEnvPrefix("CFG_BOOL")
	instance.SetEnvKeyReplacer(stringsReplacer())
	instance.AutomaticEnv()
	if err := instance.BindEnv("feature.enabled"); err != nil {
		t.Fatalf("bind env: %v", err)
	}
	missing := missingRequiredKeys(instance, []string{"feature.enabled"}, "CFG_BOOL")
	if len(missing) != 0 {
		t.Fatalf("expected no missing keys for explicit false value, got %v", missing)
	}
}

// TestParseDefaultValueFailsForInvalidValue verifies type conversion errors for malformed defaults.
func TestParseDefaultValueFailsForInvalidValue(t *testing.T) {
	_, err := parseDefaultValue("not-an-int", reflect.TypeOf(int(0)))
	if err == nil {
		t.Fatalf("expected parse default failure for invalid int")
	}
}

// TestParseDefaultValueCoversPrimitiveTypes verifies supported scalar conversions.
func TestParseDefaultValueCoversPrimitiveTypes(t *testing.T) {
	cases := []struct {
		value string
		kind  reflect.Type
	}{
		{value: "true", kind: reflect.TypeOf(true)},
		{value: "42", kind: reflect.TypeOf(int(0))},
		{value: "9", kind: reflect.TypeOf(uint(0))},
		{value: "2.5", kind: reflect.TypeOf(float64(0))},
	}
	for _, item := range cases {
		if _, err := parseDefaultValue(item.value, item.kind); err != nil {
			t.Fatalf("expected parse success for %s: %v", item.kind.Kind(), err)
		}
	}
}

// TestParseDefaultValueFailsForUnsupportedType verifies unsupported defaults fail clearly.
func TestParseDefaultValueFailsForUnsupportedType(t *testing.T) {
	_, err := parseDefaultValue("value", reflect.TypeOf([]string{}))
	if err == nil {
		t.Fatalf("expected parse failure for unsupported type")
	}
}

// TestMapKeyFallsBackToSnakeCase verifies fallback behavior for missing mapstructure tags.
func TestMapKeyFallsBackToSnakeCase(t *testing.T) {
	field := reflect.TypeOf(struct {
		SampleField string
	}{}).Field(0)
	if got := mapKey(field); got != "sample_field" {
		t.Fatalf("expected snake_case key, got %q", got)
	}
}

// TestToSnakeCaseConvertsSimpleIdentifiers verifies snake case conversion.
func TestToSnakeCaseConvertsSimpleIdentifiers(t *testing.T) {
	if got := toSnakeCase("BindIP"); got != "bind_ip" {
		t.Fatalf("expected bind_ip, got %q", got)
	}
}

// stringsReplacer returns the env key replacer used in configuration loading.
func stringsReplacer() *strings.Replacer {
	return strings.NewReplacer(".", "_")
}
