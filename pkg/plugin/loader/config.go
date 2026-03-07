package loader

import "pixelsv/pkg/plugin"

// BinaryOpener loads one plugin implementation from a shared object path.
type BinaryOpener func(path string) (plugin.Plugin, error)

// APIProvider resolves plugin APIs for enable operations.
type APIProvider func(metadata plugin.Metadata) plugin.API

// PluginState describes lifecycle state of one managed plugin.
type PluginState string

const (
	// PluginStatePending means plugin has not been enabled yet.
	PluginStatePending PluginState = "pending"
	// PluginStateSkipped means plugin was not enabled by realm or dependency gating.
	PluginStateSkipped PluginState = "skipped"
	// PluginStateEnabled means plugin OnEnable succeeded.
	PluginStateEnabled PluginState = "enabled"
	// PluginStateFailed means plugin OnEnable returned an error.
	PluginStateFailed PluginState = "failed"
	// PluginStateDisabled means plugin OnDisable already ran.
	PluginStateDisabled PluginState = "disabled"
)

// Status reports runtime state for one plugin.
type Status struct {
	// Name is the plugin name.
	Name string
	// State is the current lifecycle state.
	State PluginState
	// Err contains latest lifecycle error for the plugin.
	Err error
}

type managedPlugin struct {
	// instance is the plugin implementation instance.
	instance plugin.Plugin
	// metadata is immutable plugin metadata.
	metadata plugin.Metadata
	// enabled records whether OnEnable completed successfully.
	enabled bool
	// state records current lifecycle state.
	state PluginState
	// err records lifecycle failure details.
	err error
}
