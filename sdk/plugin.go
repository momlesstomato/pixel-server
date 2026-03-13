package sdk

// Plugin defines the contract for a loadable server extension.
type Plugin interface {
	// Manifest returns plugin metadata.
	Manifest() Manifest
	// Enable is called when the plugin is activated.
	Enable(Server) error
	// Disable is called when the server is shutting down.
	Disable() error
}

// Manifest describes plugin identity and version.
type Manifest struct {
	// Name stores plugin unique identifier.
	Name string
	// Author stores plugin author name.
	Author string
	// Version stores plugin semantic version string.
	Version string
}

// Server is the entry point for plugin interaction with the server.
type Server interface {
	// Logger returns a logger scoped to the calling plugin.
	Logger() Logger
	// Events returns the event subscription API.
	Events() EventBus
	// Sessions returns the session query and control API.
	Sessions() SessionAPI
	// Packets returns the packet send and handler registration API.
	Packets() PacketAPI
}
