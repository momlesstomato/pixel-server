package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// ApplyDefaultsFromTags sets viper defaults from `default` tags.
func ApplyDefaultsFromTags(v *viper.Viper, prefix string, sample any) error {
	typ := reflect.TypeOf(sample)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("sample must be struct, got %s", typ.Kind())
	}
	return applyDefaults(v, prefix, typ)
}

// FillDefaultsFromTags sets zero-valued fields from `default` tags.
func FillDefaultsFromTags(target any) error {
	value := reflect.ValueOf(target)
	if value.Kind() != reflect.Pointer {
		return fmt.Errorf("target must be struct pointer, got %s", value.Kind())
	}
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return fmt.Errorf("target pointer is nil")
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("target must be struct pointer, got %s", value.Kind())
	}
	return fillDefaults(value, value.Type())
}

func applyDefaults(v *viper.Viper, prefix string, typ reflect.Type) error {
	for idx := 0; idx < typ.NumField(); idx++ {
		field := typ.Field(idx)
		if !field.IsExported() {
			continue
		}
		key := mapstructureName(field)
		if key == "" {
			continue
		}
		path := joinKey(prefix, key)
		if field.Type.Kind() == reflect.Struct {
			if err := applyDefaults(v, path, field.Type); err != nil {
				return err
			}
		}
		value, ok := field.Tag.Lookup("default")
		if !ok {
			continue
		}
		parsed, err := parseDefault(field.Type.Kind(), value)
		if err != nil {
			return fmt.Errorf("invalid default for %s: %w", path, err)
		}
		v.SetDefault(path, parsed)
	}
	return nil
}

func fillDefaults(value reflect.Value, typ reflect.Type) error {
	for idx := 0; idx < typ.NumField(); idx++ {
		field := typ.Field(idx)
		if !field.IsExported() {
			continue
		}
		dst := value.Field(idx)
		if field.Type.Kind() == reflect.Struct {
			if err := fillDefaults(dst, field.Type); err != nil {
				return err
			}
		}
		defaultValue, ok := field.Tag.Lookup("default")
		if !ok || !dst.IsZero() {
			continue
		}
		parsed, err := parseDefault(field.Type.Kind(), defaultValue)
		if err != nil {
			return fmt.Errorf("invalid default for %s: %w", field.Name, err)
		}
		src := reflect.ValueOf(parsed)
		if src.Type().AssignableTo(dst.Type()) {
			dst.Set(src)
			continue
		}
		if src.Type().ConvertibleTo(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
			continue
		}
		return fmt.Errorf("default type mismatch for %s", field.Name)
	}
	return nil
}

func mapstructureName(field reflect.StructField) string {
	tag := field.Tag.Get("mapstructure")
	if tag == "" {
		return strings.ToLower(field.Name[:1]) + field.Name[1:]
	}
	name := strings.Split(tag, ",")[0]
	if name == "-" {
		return ""
	}
	return name
}

func joinKey(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func parseDefault(kind reflect.Kind, value string) (any, error) {
	switch kind {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Int:
		return strconv.Atoi(value)
	case reflect.Int32:
		parsed, err := strconv.ParseInt(value, 10, 32)
		return int32(parsed), err
	default:
		return nil, fmt.Errorf("unsupported kind %s", kind)
	}
}
