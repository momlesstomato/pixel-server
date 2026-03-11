package app

// Config defines application network and identity settings.
type Config struct {
	// BindIP sets the interface address for the server listener.
	BindIP string `mapstructure:"bind_ip" default:"0.0.0.0"`
	// Port sets the network port for the server listener.
	Port int `mapstructure:"port" default:"3000"`
	// Name sets the logical service name.
	Name string `mapstructure:"name" default:"pixel-server"`
	// Environment sets the runtime environment name.
	Environment string `mapstructure:"environment" default:"development"`
	// APIKey sets the shared key required by all API routes.
	APIKey string `mapstructure:"api_key"`
}
