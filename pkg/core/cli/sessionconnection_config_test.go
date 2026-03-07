package cli

import (
	"testing"
	"time"

	httpserver "pixelsv/pkg/http"
)

// TestSessionConnectionConfigFromHTTP validates config mapping from HTTP settings.
func TestSessionConnectionConfigFromHTTP(t *testing.T) {
	httpCfg := &httpserver.Config{
		WebSocketPingIntervalSeconds:    12,
		WebSocketPongTimeoutSeconds:     44,
		WebSocketAvailabilityOpen:       false,
		WebSocketAvailabilityOnShutdown: true,
		WebSocketAvailabilityAuthentic:  false,
		WebSocketTelemetryMinIntervalMS: 450,
	}
	cfg := sessionConnectionConfigFromHTTP(httpCfg)
	if cfg.PingInterval != 12*time.Second || cfg.PongTimeout != 44*time.Second {
		t.Fatalf("unexpected timing config: %+v", cfg)
	}
	if cfg.AvailabilityOpen || !cfg.AvailabilityOnShutdown || cfg.AvailabilityAuthentic {
		t.Fatalf("unexpected availability config: %+v", cfg)
	}
	if cfg.TelemetryMinInterval != 450*time.Millisecond {
		t.Fatalf("unexpected telemetry config: %+v", cfg)
	}
}
