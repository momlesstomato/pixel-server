package cli

import (
	"context"
	"strings"
	"time"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
	httpopenapi "github.com/momlesstomato/pixel-server/core/http/openapi"
	"github.com/momlesstomato/pixel-server/core/initializer"
	authenticationhttpapi "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/httpapi"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	managementhttpapi "github.com/momlesstomato/pixel-server/pkg/management/adapter/httpapi"
	messengerhttpapi "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/httpapi"
	messengerrealtime "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/realtime"
	permissionhttpapi "github.com/momlesstomato/pixel-server/pkg/permission/adapter/httpapi"
	userhttpapi "github.com/momlesstomato/pixel-server/pkg/user/adapter/httpapi"
	userrealtime "github.com/momlesstomato/pixel-server/pkg/user/adapter/realtime"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
)

// compositeRuntime dispatches packets to an ordered list of user runtimes.
type compositeRuntime struct {
	// runtimes stores ordered runtimes to try for each packet.
	runtimes []handshakerealtime.UserRuntime
}

// Handle tries each runtime in order until one claims the packet.
func (c *compositeRuntime) Handle(ctx context.Context, connID string, packetID uint16, payload []byte) (bool, error) {
	for _, r := range c.runtimes {
		handled, err := r.Handle(ctx, connID, packetID, payload)
		if handled || err != nil {
			return handled, err
		}
	}
	return false, nil
}

// Dispose releases all runtime resources for one connection.
func (c *compositeRuntime) Dispose(connID string) {
	for _, r := range c.runtimes {
		r.Dispose(connID)
	}
}

// registerServeWebSocket registers websocket endpoint behavior.
func registerServeWebSocket(module *corehttp.Module, path string, runtime *initializer.Runtime, services *serveServices) error {
	handler, err := handshakerealtime.NewHandler(services.sso, services.registry, packetsecurity.NewMachineIDPolicy(nil), services.bus, runtime.Logger, 30*time.Second)
	if err != nil {
		return err
	}
	handler.ConfigureBroadcaster(services.broadcaster)
	handler.ConfigureUserFinder(&userFinderAdapter{service: services.users})
	handler.ConfigurePostAuth(services.hotelStatus, services.users, services.users, services.permissions, runtime.Config.App.Name)
	handler.ConfigureUserRuntime(func(transport *handshakerealtime.Transport) (handshakerealtime.UserRuntime, error) {
		options := userrealtime.Options{
			Debounce: 2 * time.Second,
			PacketIDs: userrealtime.PacketIDs{
				SettingsRoomInvites: uint16(runtime.Config.Users.SettingsRoomInvitesPacketID),
				SettingsOldChat:     uint16(runtime.Config.Users.SettingsOldChatPacketID),
				Unignore:            uint16(runtime.Config.Users.UnignorePacketID),
				IgnoreByID:          uint16(runtime.Config.Users.IgnoreByIDPacketID),
				ApproveName:         uint16(runtime.Config.Users.ApproveNamePacketID),
			},
			Logger: runtime.Logger,
		}
		userRT, err := userrealtime.NewRuntime(services.users, services.registry, transport, options)
		if err != nil {
			return nil, err
		}
		msgRT, err := messengerrealtime.NewRuntime(services.messenger, services.registry, services.broadcaster, transport, messengerrealtime.Options{Logger: runtime.Logger})
		if err != nil {
			return nil, err
		}
		return &compositeRuntime{runtimes: []handshakerealtime.UserRuntime{userRT, msgRT}}, nil
	})
	services.handler = handler
	webSocketPath := strings.TrimSpace(path)
	if webSocketPath == "" {
		webSocketPath = "/ws"
	}
	return module.RegisterWebSocket(webSocketPath, handler.Handle)
}

// registerServeHTTPRoutes registers all REST API routes and OpenAPI documentation.
func registerServeHTTPRoutes(module *corehttp.Module, services *serveServices, wsPath string) error {
	if err := authenticationhttpapi.RegisterRoutes(module, services.sso); err != nil {
		return err
	}
	closer := &busCloserAdapter{bus: services.bus}
	if err := managementhttpapi.RegisterSessionRoutes(module, services.registry, closer); err != nil {
		return err
	}
	if err := managementhttpapi.RegisterHotelRoutes(module, services.hotelStatus); err != nil {
		return err
	}
	if err := userhttpapi.RegisterRoutes(module, services.users); err != nil {
		return err
	}
	if err := permissionhttpapi.RegisterRoutes(module, services.permissions); err != nil {
		return err
	}
	if err := messengerhttpapi.RegisterRoutes(module, services.messenger); err != nil {
		return err
	}
	paths := mergeOpenAPIPaths(
		authenticationhttpapi.OpenAPIPaths(), managementhttpapi.OpenAPIPaths(),
		userhttpapi.OpenAPIPaths(), permissionhttpapi.OpenAPIPaths(),
		messengerhttpapi.OpenAPIPaths(),
	)
	return httpopenapi.RegisterRoutes(module, httpopenapi.BuildDocument(wsPath, paths), "", "")
}

// userFinderAdapter adapts user application Service to authflow.UserFinder interface.
type userFinderAdapter struct {
	// service stores user application service behavior.
	service *userapplication.Service
}

// FindByID resolves one username by user identifier.
func (adapter *userFinderAdapter) FindByID(ctx context.Context, id int) (string, error) {
	user, err := adapter.service.FindByID(ctx, id)
	if err != nil {
		return "", err
	}
	return user.Username, nil
}

// busCloserAdapter adapts DistributedCloseSignalBus to SessionCloser interface.
type busCloserAdapter struct {
	bus *handshakerealtime.DistributedCloseSignalBus
}

// Close publishes a close signal for one connection identifier.
func (adapter *busCloserAdapter) Close(ctx context.Context, connID string, code int, reason string) error {
	return adapter.bus.Publish(ctx, connID, handshakerealtime.CloseSignal{Code: code, Reason: reason})
}

// mergeOpenAPIPaths combines multiple OpenAPI path maps.
func mergeOpenAPIPaths(maps ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, value := range maps {
		for path, pathItem := range value {
			merged[path] = pathItem
		}
	}
	return merged
}
