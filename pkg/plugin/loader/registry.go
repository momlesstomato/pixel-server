package loader

import (
	"fmt"
	"slices"

	"go.uber.org/zap"
	"pixelsv/pkg/plugin"
)

// Registry manages plugin lifecycle based on dependency and realm constraints.
type Registry struct {
	// ordered stores plugins in dependency-safe order.
	ordered []*managedPlugin
	// byName indexes managed plugins by metadata name.
	byName map[string]*managedPlugin
	// activeRealms contains enabled runtime realm names.
	activeRealms map[string]struct{}
	// logger records lifecycle outcomes.
	logger *zap.Logger
}

// New creates a plugin registry from discovered plugin instances.
func New(instances []plugin.Plugin, activeRealms []string, logger *zap.Logger) (*Registry, error) {
	orderedInstances, err := SortByDependencies(instances)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	registry := &Registry{byName: make(map[string]*managedPlugin), logger: logger}
	registry.activeRealms = toRealmSet(activeRealms)
	for _, instance := range orderedInstances {
		metadata := instance.Metadata()
		managed := &managedPlugin{instance: instance, metadata: metadata, state: PluginStatePending}
		registry.ordered = append(registry.ordered, managed)
		registry.byName[metadata.Name] = managed
	}
	return registry, nil
}

// Plugins returns plugins in dependency-safe order.
func (r *Registry) Plugins() []plugin.Plugin {
	if r == nil {
		return nil
	}
	result := make([]plugin.Plugin, 0, len(r.ordered))
	for _, managed := range r.ordered {
		result = append(result, managed.instance)
	}
	return result
}

// Status returns lifecycle state snapshots for all managed plugins.
func (r *Registry) Status() []Status {
	if r == nil {
		return nil
	}
	result := make([]Status, 0, len(r.ordered))
	for _, managed := range r.ordered {
		result = append(result, Status{Name: managed.metadata.Name, State: managed.state, Err: managed.err})
	}
	return result
}

// realmAllowed checks whether one plugin should run for active realm set.
func (r *Registry) realmAllowed(realms []string) bool {
	if len(realms) == 0 || len(r.activeRealms) == 0 {
		return true
	}
	for _, realm := range realms {
		if _, ok := r.activeRealms[realm]; ok {
			return true
		}
	}
	return false
}

// dependenciesReady validates that all dependencies are enabled.
func (r *Registry) dependenciesReady(dependencies []string) error {
	for _, dep := range dependencies {
		managed, ok := r.byName[dep]
		if !ok {
			return fmt.Errorf("dependency %s is not loaded", dep)
		}
		if !managed.enabled {
			return fmt.Errorf("dependency %s is not enabled", dep)
		}
	}
	return nil
}

// toRealmSet converts role list to an indexed realm set.
func toRealmSet(realms []string) map[string]struct{} {
	set := make(map[string]struct{}, len(realms))
	for _, realm := range slices.Compact(realms) {
		if realm != "" {
			set[realm] = struct{}{}
		}
	}
	return set
}
