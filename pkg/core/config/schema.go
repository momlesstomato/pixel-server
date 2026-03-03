package config

import (
	"fmt"
	"reflect"
)

// FieldMeta describes one configuration field discovered from struct tags.
type FieldMeta struct {
	// Key is the viper key, usually from mapstructure tag.
	Key string

	// Env is the environment variable name from env tag.
	Env string

	// Default is the default value from default tag.
	Default string

	// HasDefault indicates whether a default tag was set.
	HasDefault bool
}

func collectFieldMeta(t reflect.Type) ([]FieldMeta, error) {
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("config target must be pointer to struct")
	}
	var out []FieldMeta
	walkFields(t.Elem(), "", &out)
	return out, nil
}

func walkFields(t reflect.Type, prefix string, out *[]FieldMeta) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			nestedPrefix := prefix
			if tagKey := field.Tag.Get("mapstructure"); tagKey != "" {
				if nestedPrefix == "" {
					nestedPrefix = tagKey
				} else {
					nestedPrefix = nestedPrefix + "." + tagKey
				}
			}
			walkFields(field.Type, nestedPrefix, out)
			continue
		}

		key := field.Tag.Get("mapstructure")
		env := field.Tag.Get("env")
		defaultValue, hasDefault := field.Tag.Lookup("default")
		if key == "" {
			continue
		}
		if prefix != "" {
			key = prefix + "." + key
		}
		*out = append(*out, FieldMeta{
			Key:        key,
			Env:        env,
			Default:    defaultValue,
			HasDefault: hasDefault,
		})
	}
}
