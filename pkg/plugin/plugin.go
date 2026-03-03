package plugin

// Meta contains static information about a plugin.
// It is returned by Plugin.Meta() and is used by the Registry to resolve
// dependency ordering and display information to operators.
type Meta struct {
	// Name is the unique identifier for the plugin (no spaces).
	Name string

	// Version is the semver string, e.g. "1.0.0".
	Version string

	// Author is the creator of the plugin.
	Author string

	// Description is a short human-readable description.
	Description string

	// Depends lists the Names of other plugins that must be enabled before
	// this one. The Registry uses this for topological ordering.
	Depends []string
}

// Plugin is the interface that every pixel-server plugin must implement.
// The shared object (.so) must export a NewPlugin symbol:
//
//	func NewPlugin() Plugin
type Plugin interface {
	// Meta returns static metadata. Called once during loading.
	Meta() Meta

	// OnEnable is called after the plugin is loaded and its dependencies are
	// already enabled. api gives scoped access to the server facilities.
	// Return non-nil to abort loading this plugin (it will be skipped).
	OnEnable(api API) error

	// OnDisable is called during a graceful shutdown, in reverse-dependency
	// order. Plugins must release any resources they hold.
	OnDisable() error
}
