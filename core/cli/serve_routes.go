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
	permissionhttpapi "github.com/momlesstomato/pixel-server/pkg/permission/adapter/httpapi"
	userhttpapi "github.com/momlesstomato/pixel-server/pkg/user/adapter/httpapi"
	userrealtime "github.com/momlesstomato/pixel-server/pkg/user/adapter/realtime"
)

// registerServeWebSocket registers websocket endpoint behavior.
func registerServeWebSocket(module *corehttp.Module, path string, runtime *initializer.Runtime, services *serveServices) error {
	handler, err := handshakerealtime.NewHandler(services.sso, services.registry, packetsecurity.NewMachineIDPolicy(nil), services.bus, runtime.Logger, 30*time.Second)
	if err != nil {
		return err
	}
	handler.ConfigureBroadcaster(services.broadcaster)
	handler.ConfigurePostAuth(services.hotelStatus, services.users, services.users, services.permissions, runtime.Config.App.Name)
	handler.SetUserFinder(services)
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
		return userrealtime.NewRuntime(services.users, services.registry, transport, options)
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
	paths := mergeOpenAPIPaths(authenticationhttpapi.OpenAPIPaths(), managementhttpapi.OpenAPIPaths(), userhttpapi.OpenAPIPaths(), permissionhttpapi.OpenAPIPaths())
	return httpopenapi.RegisterRoutes(module, httpopenapi.BuildDocument(wsPath, paths), "", "")
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
