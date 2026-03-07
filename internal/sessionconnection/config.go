package sessionconnection

import (
	"fmt"
	"time"
)

// Config defines session-connection realm runtime settings.
type Config struct {
	// PingInterval defines keepalive ping interval.
	PingInterval time.Duration
	// PongTimeout defines max time without client pong before disconnect.
	PongTimeout time.Duration
	// AvailabilityOpen controls availability.status isOpen flag.
	AvailabilityOpen bool
	// AvailabilityOnShutdown controls availability.status onShutdown flag.
	AvailabilityOnShutdown bool
	// AvailabilityAuthentic controls availability.status isAuthentic flag.
	AvailabilityAuthentic bool
	// TelemetryMinInterval defines minimum interval between telemetry logs per packet.
	TelemetryMinInterval time.Duration
}

// DefaultConfig returns default session-connection settings.
func DefaultConfig() Config {
	return Config{
		PingInterval:           30 * time.Second,
		PongTimeout:            90 * time.Second,
		AvailabilityOpen:       true,
		AvailabilityOnShutdown: false,
		AvailabilityAuthentic:  true,
		TelemetryMinInterval:   time.Second,
	}
}

// Validate checks whether Config values are valid.
func (c Config) Validate() error {
	if c.PingInterval < time.Second {
		return fmt.Errorf("invalid ping interval: %s", c.PingInterval)
	}
	if c.PongTimeout < c.PingInterval {
		return fmt.Errorf("invalid pong timeout: %s", c.PongTimeout)
	}
	if c.TelemetryMinInterval < 0 {
		return fmt.Errorf("invalid telemetry min interval: %s", c.TelemetryMinInterval)
	}
	return nil
}
