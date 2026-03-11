package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/viper"
)

// bindDefaultsAndEnv registers defaults and environment bindings for each leaf config field.
func bindDefaultsAndEnv(instance *viper.Viper, kind reflect.Type, prefix string) ([]string, []string, error) {
	keys := make([]string, 0, 16)
	required := make([]string, 0, 8)
	for index := 0; index < kind.NumField(); index++ {
		field := kind.Field(index)
		key := mapKey(field)
		if key == "-" {
			continue
		}
		full := key
		if prefix != "" {
			full = prefix + "." + key
		}
		if field.Type.Kind() == reflect.Struct {
			nestedKeys, nestedRequired, err := bindDefaultsAndEnv(instance, field.Type, full)
			if err != nil {
				return nil, nil, err
			}
			keys = append(keys, nestedKeys...)
			required = append(required, nestedRequired...)
			continue
		}
		if err := instance.BindEnv(full); err != nil {
			return nil, nil, fmt.Errorf("bind env for %q: %w", full, err)
		}
		keys = append(keys, full)
		defaultValue, hasDefault := field.Tag.Lookup("default")
		if !hasDefault {
			required = append(required, full)
			continue
		}
		parsed, err := parseDefaultValue(defaultValue, field.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("parse default for %q: %w", full, err)
		}
		instance.SetDefault(full, parsed)
	}
	return keys, required, nil
}

// applyAliasValues maps env-style keys from .env files into nested configuration keys.
func applyAliasValues(instance *viper.Viper, keys []string, prefix string) {
	for _, key := range keys {
		if _, exists := os.LookupEnv(envVariableName(key, prefix)); exists {
			continue
		}
		upper := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
		lower := strings.ToLower(upper)
		if instance.IsSet(upper) {
			instance.Set(key, instance.Get(upper))
			continue
		}
		if instance.IsSet(lower) {
			instance.Set(key, instance.Get(lower))
		}
	}
}

// missingRequiredKeys returns the required keys that were not provided by any source.
func missingRequiredKeys(instance *viper.Viper, required []string, prefix string) []string {
	missing := make([]string, 0, len(required))
	for _, key := range required {
		if instance.IsSet(key) {
			continue
		}
		if _, exists := os.LookupEnv(envVariableName(key, prefix)); exists {
			continue
		}
		missing = append(missing, envVariableName(key, prefix))
	}
	return missing
}

// envVariableName converts a nested key into uppercase env var name.
func envVariableName(key, prefix string) string {
	normalized := strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if prefix == "" {
		return normalized
	}
	return strings.ToUpper(prefix) + "_" + normalized
}

// mapKey extracts the canonical mapstructure key for a field.
func mapKey(field reflect.StructField) string {
	tag := field.Tag.Get("mapstructure")
	if tag == "" {
		return toSnakeCase(field.Name)
	}
	return strings.Split(tag, ",")[0]
}

// parseDefaultValue converts default tag content to the target field type.
func parseDefaultValue(value string, kind reflect.Type) (any, error) {
	switch kind.Kind() {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.ParseInt(value, 10, int(kind.Bits()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.ParseUint(value, 10, int(kind.Bits()))
	case reflect.Float32, reflect.Float64:
		return strconv.ParseFloat(value, int(kind.Bits()))
	default:
		return nil, fmt.Errorf("unsupported default type %s", kind.Kind())
	}
}

// toSnakeCase converts PascalCase identifiers to snake_case.
func toSnakeCase(value string) string {
	var out []rune
	for idx, current := range value {
		if idx > 0 && unicode.IsUpper(current) && !unicode.IsUpper(rune(value[idx-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(current))
	}
	return string(out)
}
