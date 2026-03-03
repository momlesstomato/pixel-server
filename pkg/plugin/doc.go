// Package plugin provides the plugin framework for pixel-server.
//
// The framework is split into focused sub-packages:
//   - plugin/event:     in-process synchronous event bus
//   - plugin/intercept: packet interception pipeline (before/after hooks)
//   - plugin/roomsvc:   ECS-safe room facade for plugins
//
// This root package contains the Plugin interface, Meta, API, Registry,
// and loader. Plugin .so files import this package.
package plugin
