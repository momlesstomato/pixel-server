package loader

import (
	"fmt"
	"sort"

	"pixelsv/pkg/plugin"
)

// SortByDependencies topologically orders plugins by DependsOn metadata.
func SortByDependencies(instances []plugin.Plugin) ([]plugin.Plugin, error) {
	nodes, err := buildNodes(instances)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	names := sortedKeys(nodes)
	mark := make(map[string]uint8, len(nodes))
	ordered := make([]string, 0, len(nodes))
	for _, name := range names {
		if err := visit(name, nodes, mark, &ordered); err != nil {
			return nil, err
		}
	}
	result := make([]plugin.Plugin, 0, len(ordered))
	for _, name := range ordered {
		result = append(result, nodes[name].instance)
	}
	return result, nil
}

// dependencyNode groups one plugin instance with metadata for sorting.
type dependencyNode struct {
	// instance is the plugin object.
	instance plugin.Plugin
	// metadata is immutable plugin metadata.
	metadata plugin.Metadata
}

// buildNodes validates metadata and indexes plugins by name.
func buildNodes(instances []plugin.Plugin) (map[string]dependencyNode, error) {
	nodes := make(map[string]dependencyNode, len(instances))
	for _, instance := range instances {
		if instance == nil {
			return nil, fmt.Errorf("plugin instance is nil")
		}
		metadata := instance.Metadata()
		if metadata.Name == "" {
			return nil, fmt.Errorf("plugin metadata name is required")
		}
		if _, exists := nodes[metadata.Name]; exists {
			return nil, fmt.Errorf("duplicate plugin name: %s", metadata.Name)
		}
		nodes[metadata.Name] = dependencyNode{instance: instance, metadata: metadata}
	}
	return nodes, nil
}

// visit performs DFS topological ordering with cycle checks.
func visit(name string, nodes map[string]dependencyNode, mark map[string]uint8, ordered *[]string) error {
	if mark[name] == 2 {
		return nil
	}
	if mark[name] == 1 {
		return fmt.Errorf("plugin dependency cycle detected at %s", name)
	}
	mark[name] = 1
	deps := append([]string(nil), nodes[name].metadata.DependsOn...)
	sort.Strings(deps)
	for _, dep := range deps {
		if _, exists := nodes[dep]; !exists {
			return fmt.Errorf("plugin %s depends on missing plugin %s", name, dep)
		}
		if err := visit(dep, nodes, mark, ordered); err != nil {
			return err
		}
	}
	mark[name] = 2
	*ordered = append(*ordered, name)
	return nil
}

// sortedKeys returns deterministic sorted map keys.
func sortedKeys(nodes map[string]dependencyNode) []string {
	keys := make([]string, 0, len(nodes))
	for name := range nodes {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
