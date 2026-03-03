package plugin

func (r *Registry) sortByDependencies() {
	indexByName := make(map[string]int, len(r.plugins))
	for i, pluginEntry := range r.plugins {
		indexByName[pluginEntry.meta.Name] = i
	}

	visited := make(map[string]bool, len(r.plugins))
	order := make([]*entry, 0, len(r.plugins))

	var visit func(pluginEntry *entry)
	visit = func(pluginEntry *entry) {
		if visited[pluginEntry.meta.Name] {
			return
		}
		visited[pluginEntry.meta.Name] = true
		for _, dependency := range pluginEntry.meta.Depends {
			if i, ok := indexByName[dependency]; ok {
				visit(r.plugins[i])
			}
		}
		order = append(order, pluginEntry)
	}

	for _, pluginEntry := range r.plugins {
		visit(pluginEntry)
	}
	r.plugins = order
}
