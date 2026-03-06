package factory

import (
	"errors"

	"github.com/nats-io/nats.go"
	"pixelsv/pkg/core/transport"
	"pixelsv/pkg/core/transport/local"
	natsbus "pixelsv/pkg/core/transport/nats"
)

// Config controls runtime transport adapter selection.
type Config struct {
	// NATSURL selects NATS transport when set and not forced local.
	NATSURL string `default:""`
	// ForceLocal enforces in-process transport regardless of NATS URL.
	ForceLocal bool `default:"false"`
}

// New creates a transport bus according to Config.
func New(cfg Config, options ...nats.Option) (transport.Bus, error) {
	if cfg.ForceLocal || cfg.NATSURL == "" {
		return local.New(), nil
	}
	bus, err := natsbus.New(cfg.NATSURL, options...)
	if err != nil {
		return nil, err
	}
	if bus == nil {
		return nil, errors.New("transport bus is nil")
	}
	return bus, nil
}
