package config

import (
	"fmt"
	"strings"
)

// Config defines the root configuration for shared runtime concerns.
type Config struct {
	// App contains generic application settings.
	App AppConfig `mapstructure:"app"`
	// Runtime contains runtime role and transport settings.
	Runtime RuntimeConfig `mapstructure:"runtime"`
}

// AppConfig contains base application settings.
type AppConfig struct {
	// Env identifies the runtime environment.
	Env string `mapstructure:"env" default:"development"`
}

// RuntimeConfig contains process-level runtime settings.
type RuntimeConfig struct {
	// Role defines active runtime roles as a comma-separated string.
	Role string `mapstructure:"role" default:"all"`
	// InstanceID identifies the runtime process instance.
	InstanceID string `mapstructure:"instance_id" default:"pixelsv-local"`
	// NATSURL defines optional NATS transport URL.
	NATSURL string `mapstructure:"nats_url" default:""`
}

// Validate checks whether Config is internally consistent.
func (c Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("app: %w", err)
	}
	if err := c.Runtime.Validate(); err != nil {
		return fmt.Errorf("runtime: %w", err)
	}
	return nil
}

// Validate checks whether AppConfig is internally consistent.
func (c AppConfig) Validate() error {
	if c.Env == "" {
		return ErrEmptyAppEnv
	}
	return nil
}

// Validate checks whether RuntimeConfig is internally consistent.
func (c RuntimeConfig) Validate() error {
	if c.Role == "" {
		return ErrEmptyRuntimeRole
	}
	if c.InstanceID == "" {
		return ErrEmptyRuntimeInstanceID
	}
	if _, err := ParseRoles(c.Role); err != nil {
		return err
	}
	return nil
}

// ParseRoles normalizes and validates a comma-separated role string.
func ParseRoles(value string) ([]string, error) {
	tokens := strings.Split(value, ",")
	roles := make([]string, 0, len(tokens))
	seen := map[string]struct{}{}
	hasAll := false
	for _, token := range tokens {
		role := strings.ToLower(strings.TrimSpace(token))
		if role == "" {
			continue
		}
		if role == "all" {
			hasAll = true
			continue
		}
		if _, ok := validRuntimeRoles[role]; !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidRuntimeRole, role)
		}
		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	if hasAll {
		return []string{"all"}, nil
	}
	if len(roles) == 0 {
		return nil, ErrEmptyRuntimeRole
	}
	return roles, nil
}

var validRuntimeRoles = map[string]struct{}{
	"all":        {},
	"gateway":    {},
	"game":       {},
	"auth":       {},
	"social":     {},
	"navigator":  {},
	"catalog":    {},
	"moderation": {},
	"api":        {},
	"jobs":       {},
}
