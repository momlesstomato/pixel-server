package cli

import (
	"time"

	"pixelsv/internal/sessionconnection"
	httpserver "pixelsv/pkg/http"
)

// sessionConnectionConfigFromHTTP maps HTTP settings to session-connection runtime config.
func sessionConnectionConfigFromHTTP(httpCfg *httpserver.Config) sessionconnection.Config {
	cfg := sessionconnection.DefaultConfig()
	if httpCfg == nil {
		return cfg
	}
	cfg.PingInterval = time.Duration(httpCfg.WebSocketPingIntervalSeconds) * time.Second
	cfg.PongTimeout = time.Duration(httpCfg.WebSocketPongTimeoutSeconds) * time.Second
	cfg.AvailabilityOpen = httpCfg.WebSocketAvailabilityOpen
	cfg.AvailabilityOnShutdown = httpCfg.WebSocketAvailabilityOnShutdown
	cfg.AvailabilityAuthentic = httpCfg.WebSocketAvailabilityAuthentic
	cfg.TelemetryMinInterval = time.Duration(httpCfg.WebSocketTelemetryMinIntervalMS) * time.Millisecond
	return cfg
}
