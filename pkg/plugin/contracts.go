package plugin

// Metadata describes immutable plugin identity and lifecycle constraints.
type Metadata struct {
	// Name is the unique plugin identifier.
	Name string
	// Version is the semantic version of the plugin implementation.
	Version string
	// Realms lists target realms; empty means all active realms.
	Realms []string
	// DependsOn lists plugin names that must be enabled first.
	DependsOn []string
}

// Scope describes host runtime identity exposed to plugins.
type Scope struct {
	// InstanceID identifies the current pixelsv runtime instance.
	InstanceID string
	// Version identifies the host pixelsv build version.
	Version string
	// Environment identifies host environment labels such as dev or prod.
	Environment string
}

// Plugin defines the lifecycle contract for runtime extensions.
type Plugin interface {
	// Metadata returns immutable plugin identity and dependency metadata.
	Metadata() Metadata
	// OnEnable initializes plugin hooks and resources.
	OnEnable(api API) error
	// OnDisable releases plugin resources and unregisters runtime hooks.
	OnDisable() error
}
