package main

import sdk "github.com/momlesstomato/pixel-sdk"

// loginLogger logs authentication events using the plugin logger.
type loginLogger struct {
	server sdk.Server
}

// Manifest returns plugin identity metadata.
func (p *loginLogger) Manifest() sdk.Manifest {
	return sdk.Manifest{Name: "login-logger", Author: "pixelsv", Version: "1.0.0"}
}

// Enable subscribes to auth lifecycle events.
func (p *loginLogger) Enable(server sdk.Server) error {
	p.server = server
	server.Events().Subscribe(func(e *sdk.AuthCompleted) {
		server.Logger().Printf("user %d authenticated on connection %s", e.UserID, e.ConnID)
	})
	server.Events().Subscribe(func(e *sdk.ConnectionClosed) {
		server.Logger().Printf("connection %s closed", e.ConnID)
	})
	server.Events().Subscribe(func(e *sdk.DuplicateKick) {
		server.Logger().Printf("duplicate login for user %d: kicking %s in favor of %s", e.UserID, e.OldConnID, e.NewConnID)
	})
	server.Logger().Printf("login-logger plugin enabled")
	return nil
}

// Disable cleans up plugin resources.
func (p *loginLogger) Disable() error {
	p.server.Logger().Printf("login-logger plugin disabled")
	return nil
}

// NewPlugin is the exported symbol used by the .so loader.
func NewPlugin() sdk.Plugin {
	return &loginLogger{}
}

func main() {}
