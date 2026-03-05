package log

import (
	"reflect"
	"testing"
)

// TestConfigValidate checks config validation rules.
func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{name: "valid console", cfg: Config{Format: FormatConsole, Level: "info"}, wantErr: false},
		{name: "valid json", cfg: Config{Format: FormatJSON, Level: "debug"}, wantErr: false},
		{name: "invalid format", cfg: Config{Format: "xml", Level: "info"}, wantErr: true},
		{name: "invalid level", cfg: Config{Format: FormatJSON, Level: "bad"}, wantErr: true},
		{name: "empty level", cfg: Config{Format: FormatJSON, Level: ""}, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

// TestZapConfig validates zap config building.
func TestZapConfig(t *testing.T) {
	cfg := Config{Format: FormatJSON, Level: "warn"}
	zapCfg, err := ZapConfig(cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if zapCfg.Encoding != FormatJSON {
		t.Fatalf("expected json encoding")
	}
}

// TestNewInvalidConfig validates logger construction failure path.
func TestNewInvalidConfig(t *testing.T) {
	if _, err := New(Config{Format: "xml", Level: "info"}); err == nil {
		t.Fatalf("expected validation error")
	}
}

// TestConfigDefaultTags checks default tag presence.
func TestConfigDefaultTags(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	if typ.Field(0).Tag.Get("default") == "" || typ.Field(1).Tag.Get("default") == "" {
		t.Fatalf("expected default tags on log config fields")
	}
}
