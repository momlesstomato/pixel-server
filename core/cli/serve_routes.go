package cli

import (
	"context"
	"strings"
	"time"

	corehttp "github.com/momlesstomato/pixel-server/core/http"
	httpopenapi "github.com/momlesstomato/pixel-server/core/http/openapi"
	"github.com/momlesstomato/pixel-server/core/initializer"
	authenticationhttpapi "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/httpapi"
	cataloghttpapi "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/httpapi"
	economyhttpapi "github.com/momlesstomato/pixel-server/pkg/economy/adapter/httpapi"
	furniturehttpapi "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/httpapi"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	inventoryhttpapi "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/httpapi"
	managementhttpapi "github.com/momlesstomato/pixel-server/pkg/management/adapter/httpapi"
	messengerhttpapi "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/httpapi"
	messengerrealtime "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/realtime"
	navigatorhttpapi "github.com/momlesstomato/pixel-server/pkg/navigator/adapter/httpapi"
	permissionhttpapi "github.com/momlesstomato/pixel-server/pkg/permission/adapter/httpapi"
	roomrealtime "github.com/momlesstomato/pixel-server/pkg/room/adapter/realtime"
	subscriptionhttpapi "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/httpapi"
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
		runtimes := []handshakerealtime.UserRuntime{userRT, msgRT}
		ecoRTs, err := buildEconomyRuntimes(services.economyBundle, services.registry, transport, runtime.Logger)
		if err != nil {
			return nil, err
		}
		runtimes = append(runtimes, ecoRTs...)
		roomRT, err := roomrealtime.NewRuntime(services.room, services.entityService, services.chatService, services.registry, transport, runtime.Logger)
		if err != nil {
			return nil, err
		}
		services.room.Manager().SetBroadcaster(roomRT.Broadcast)
		runtimes = append(runtimes, roomRT)
		return &compositeRuntime{runtimes: runtimes}, nil
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
	closer := &busCloserAdapter{bus: services.bus}
	for _, register := range []func(*corehttp.Module) error{
		func(m *corehttp.Module) error { return authenticationhttpapi.RegisterRoutes(m, services.sso) },
		func(m *corehttp.Module) error {
			return managementhttpapi.RegisterSessionRoutes(m, services.registry, closer, services.fireSafe)
		},
		func(m *corehttp.Module) error {
			return managementhttpapi.RegisterHotelRoutes(m, services.hotelStatus, services.fireSafe)
		},
		func(m *corehttp.Module) error { return userhttpapi.RegisterRoutes(m, services.users) },
		func(m *corehttp.Module) error { return permissionhttpapi.RegisterRoutes(m, services.permissions) },
		func(m *corehttp.Module) error { return messengerhttpapi.RegisterRoutes(m, services.messenger) },
		func(m *corehttp.Module) error { return furniturehttpapi.RegisterRoutes(m, services.furniture) },
		func(m *corehttp.Module) error { return inventoryhttpapi.RegisterRoutes(m, services.inventory) },
		func(m *corehttp.Module) error { return cataloghttpapi.RegisterRoutes(m, services.catalog) },
		func(m *corehttp.Module) error { return economyhttpapi.RegisterRoutes(m, services.economy) },
		func(m *corehttp.Module) error { return subscriptionhttpapi.RegisterRoutes(m, services.subscription) },
		func(m *corehttp.Module) error { return navigatorhttpapi.RegisterRoutes(m, services.navigator) },
	} {
		if err := register(module); err != nil {
			return err
		}
	}
	paths := mergeOpenAPIPaths(
		authenticationhttpapi.OpenAPIPaths(), managementhttpapi.OpenAPIPaths(),
		userhttpapi.OpenAPIPaths(), permissionhttpapi.OpenAPIPaths(),
		messengerhttpapi.OpenAPIPaths(), furniturehttpapi.OpenAPIPaths(),
		inventoryhttpapi.OpenAPIPaths(), cataloghttpapi.OpenAPIPaths(),
		economyhttpapi.OpenAPIPaths(), subscriptionhttpapi.OpenAPIPaths(), navigatorhttpapi.OpenAPIPaths(),
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
