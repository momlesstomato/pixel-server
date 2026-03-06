package httpserver

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
)

// TestFromViper validates env binding and config parsing.
func TestFromViper(t *testing.T) {
	t.Setenv("API_KEY", "secret")
	t.Setenv("HTTP_ADDR", ":9090")
	v := viper.New()
	if err := BindViper(v); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	cfg, err := FromViper(v)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Address != ":9090" || cfg.APIKey != "secret" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

// TestConfigValidate checks required API key validation.
func TestConfigValidate(t *testing.T) {
	if err := (Config{Address: ":8080", APIKey: "", ReadTimeoutSeconds: 10}).Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
	if err := (Config{Address: ":8080", APIKey: "secret", ReadTimeoutSeconds: 0}).Validate(); err == nil {
		t.Fatalf("expected invalid timeout error")
	}
}

// TestConfigDefaultTags validates default tags contract.
func TestConfigDefaultTags(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	apiKeyField, _ := typ.FieldByName("APIKey")
	if apiKeyField.Tag.Get("default") != "" {
		t.Fatalf("expected APIKey to be required without default tag")
	}
	addressField, _ := typ.FieldByName("Address")
	if addressField.Tag.Get("default") == "" {
		t.Fatalf("expected Address default tag")
	}
}
