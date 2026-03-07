package loader

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
)

// EnableAll tries to enable every plugin while isolating plugin failures.
func (r *Registry) EnableAll(provider APIProvider) error {
	if r == nil {
		return nil
	}
	var errs []error
	for _, managed := range r.ordered {
		if !r.realmAllowed(managed.metadata.Realms) {
			managed.state = PluginStateSkipped
			managed.err = nil
			continue
		}
		if depErr := r.dependenciesReady(managed.metadata.DependsOn); depErr != nil {
			managed.state = PluginStateSkipped
			managed.err = depErr
			errs = append(errs, depErr)
			continue
		}
		if provider == nil {
			managed.state = PluginStateFailed
			managed.err = fmt.Errorf("plugin API provider is nil")
			errs = append(errs, managed.err)
			continue
		}
		enableErr := managed.instance.OnEnable(provider(managed.metadata))
		if enableErr != nil {
			managed.state = PluginStateFailed
			managed.err = enableErr
			errs = append(errs, fmt.Errorf("plugin %s enable: %w", managed.metadata.Name, enableErr))
			r.logger.Error("plugin enable failed", zap.String("plugin", managed.metadata.Name), zap.Error(enableErr))
			continue
		}
		managed.enabled = true
		managed.state = PluginStateEnabled
		managed.err = nil
	}
	return errors.Join(errs...)
}

// DisableAll disables enabled plugins in reverse dependency order.
func (r *Registry) DisableAll() error {
	if r == nil {
		return nil
	}
	var errs []error
	for index := len(r.ordered) - 1; index >= 0; index-- {
		managed := r.ordered[index]
		if !managed.enabled {
			continue
		}
		disableErr := managed.instance.OnDisable()
		managed.enabled = false
		managed.state = PluginStateDisabled
		managed.err = disableErr
		if disableErr != nil {
			errs = append(errs, fmt.Errorf("plugin %s disable: %w", managed.metadata.Name, disableErr))
		}
	}
	return errors.Join(errs...)
}
