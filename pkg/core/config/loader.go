package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

// Load reads environment values into cfg using schema tags and validates required fields.
func Load(cfg any) error {
	if cfg == nil {
		return fmt.Errorf("config target is nil")
	}
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := readDotEnv(v); err != nil {
		return err
	}

	metas, err := collectFieldMeta(reflect.TypeOf(cfg))
	if err != nil {
		return err
	}
	for _, meta := range metas {
		if meta.Env != "" && v.InConfig(meta.Env) {
			v.Set(meta.Key, v.Get(meta.Env))
		}
		if meta.HasDefault {
			v.SetDefault(meta.Key, meta.Default)
		}
		if meta.Env != "" {
			if bindErr := v.BindEnv(meta.Key, meta.Env); bindErr != nil {
				return fmt.Errorf("bind env %s to %s: %w", meta.Env, meta.Key, bindErr)
			}
		}
	}

	if err := v.Unmarshal(cfg); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	if err := validateRequired(v, metas); err != nil {
		return err
	}
	return nil
}

func readDotEnv(v *viper.Viper) error {
	for _, candidate := range []string{".env", "../.env", "../../.env"} {
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		v.SetConfigFile(candidate)
		v.SetConfigType("env")
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("read %s: %w", candidate, err)
		}
		return nil
	}
	return nil
}

func validateRequired(v *viper.Viper, metas []FieldMeta) error {
	var missing []string
	for _, meta := range metas {
		if meta.HasDefault || meta.Env == "" || meta.Key == "" {
			continue
		}
		if !v.IsSet(meta.Key) {
			missing = append(missing, meta.Env)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return nil
}
