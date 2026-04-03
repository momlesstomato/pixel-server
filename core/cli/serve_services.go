package cli

import (
	"context"
	"time"

	sdk "github.com/momlesstomato/pixel-sdk"
	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/core/initializer"
	authenticationapplication "github.com/momlesstomato/pixel-server/pkg/authentication/application"
	authenticationredisstore "github.com/momlesstomato/pixel-server/pkg/authentication/infrastructure/redisstore"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	economyapplication "github.com/momlesstomato/pixel-server/pkg/economy/application"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	inventorydomain "github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	messengerapplication "github.com/momlesstomato/pixel-server/pkg/messenger/application"
	messengerstore "github.com/momlesstomato/pixel-server/pkg/messenger/infrastructure/store"
	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	permissionnotification "github.com/momlesstomato/pixel-server/pkg/permission/adapter/notification"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissionstore "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/store"
	roomapplication "github.com/momlesstomato/pixel-server/pkg/room/application"
	sessionhotelstatus "github.com/momlesstomato/pixel-server/pkg/status/application/hotelstatus"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	subscriptionapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
)

// serveServices holds shared dependencies built during serve startup.
type serveServices struct {
	sso           *authenticationapplication.Service
	registry      *coreconnection.RedisSessionRegistry
	bus           *handshakerealtime.DistributedCloseSignalBus
	broadcaster   broadcast.Broadcaster
	hotelStatus   *sessionhotelstatus.Service
	users         *userapplication.Service
	permissions   *permissionapplication.Service
	messenger     *messengerapplication.Service
	furniture     *furnitureapplication.Service
	inventory     *inventoryapplication.Service
	catalog       *catalogapplication.Service
	economy       *economyapplication.Service
	subscription  *subscriptionapplication.Service
	navigator     *navigatorapplication.Service
	room          *roomapplication.Service
	entityService *roomapplication.EntityService
	chatService   *roomapplication.ChatService
	economyBundle *economyServiceBundle
	handler       *handshakerealtime.Handler
	fire          func(sdk.Event)
}

// buildServeServices constructs shared application dependencies.
func buildServeServices(runtime *initializer.Runtime) (*serveServices, error) {
	ssoStore, err := authenticationredisstore.NewRedisStore(runtime.Redis, runtime.Config.Authentication.KeyPrefix)
	if err != nil {
		return nil, err
	}
	registry, err := coreconnection.NewRedisSessionRegistryWithOptions(runtime.Redis, coreconnection.RedisSessionRegistryOptions{InstanceID: runtime.Config.App.Name})
	if err != nil {
		return nil, err
	}
	bus, err := handshakerealtime.NewRedisCloseSignalBus(runtime.Redis, "handshake:close")
	if err != nil {
		return nil, err
	}
	broadcaster, err := broadcast.NewRedisBroadcaster(runtime.Redis, "")
	if err != nil {
		return nil, err
	}
	statusStore, err := statusredisstore.NewStore(runtime.Redis, runtime.Config.Status.RedisKey)
	if err != nil {
		return nil, err
	}
	hotelStatus, err := sessionhotelstatus.NewService(statusStore, broadcaster, runtime.Config.Status)
	if err != nil {
		return nil, err
	}
	if _, err = hotelStatus.Current(context.Background()); err != nil {
		return nil, err
	}
	hotelStatus.StartCountdownTicker(context.Background())
	userRepository, err := userstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	users, err := userapplication.NewService(userRepository)
	if err != nil {
		return nil, err
	}
	permissionRepository, err := permissionstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	permissions, err := permissionapplication.NewService(permissionRepository, runtime.Redis, permissionapplication.Config{
		CachePrefix: runtime.Config.Permission.CachePrefix, CacheTTL: time.Duration(runtime.Config.Permission.CacheTTLSeconds) * time.Second,
		AmbassadorPermission: runtime.Config.Permission.AmbassadorPermission,
	})
	if err != nil {
		return nil, err
	}
	liveUpdater, err := permissionnotification.NewLiveUpdater(broadcaster)
	if err != nil {
		return nil, err
	}
	permissions.SetLiveUpdater(liveUpdater)
	messengerRepository, err := messengerstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	messenger, err := messengerapplication.NewService(messengerRepository, registry, broadcaster, runtime.Config.Messenger)
	if err != nil {
		return nil, err
	}
	messenger.StartPurgeTicker(context.Background())
	economyServices, err := buildEconomyServices(runtime)
	if err != nil {
		return nil, err
	}
	roomService, err := buildRoomServices(runtime, noopEntityBroadcaster)
	if err != nil {
		return nil, err
	}
	entityService, err := roomapplication.NewEntityService(roomService.Manager(), runtime.Logger)
	if err != nil {
		return nil, err
	}
	chatService, err := roomapplication.NewChatService(runtime.Logger)
	if err != nil {
		return nil, err
	}
	return &serveServices{
		sso:      authenticationapplication.NewService(ssoStore, runtime.Config.Authentication),
		registry: registry, bus: bus, broadcaster: broadcaster, hotelStatus: hotelStatus,
		users: users, permissions: permissions, messenger: messenger,
		furniture: economyServices.furniture, inventory: economyServices.inventory,
		catalog: economyServices.catalog, economy: economyServices.economy,
		subscription: economyServices.subscription, navigator: economyServices.navigator,
		room: roomService, entityService: entityService, chatService: chatService,
		economyBundle: economyServices,
	}, nil
}

// fireSafe dispatches one event when the plugin dispatcher is available.
func (s *serveServices) fireSafe(event sdk.Event) {
	if s.fire != nil {
		s.fire(event)
	}
}

// inventorySpender adapts inventory.Service to catalogdomain.Spender.
type inventorySpender struct {
	// svc stores inventory application service.
	svc *inventoryapplication.Service
}

// GetCredits delegates to inventory credit balance lookup.
func (s *inventorySpender) GetCredits(ctx context.Context, userID int) (int, error) {
	return s.svc.GetCredits(ctx, userID)
}

// AddCredits delegates to inventory credit adjustment.
func (s *inventorySpender) AddCredits(ctx context.Context, userID int, amount int) (int, error) {
	return s.svc.AddCredits(ctx, userID, amount)
}

// GetCurrencyBalance delegates to inventory activity-point balance lookup.
func (s *inventorySpender) GetCurrencyBalance(ctx context.Context, userID int, typeID int) (int, error) {
	return s.svc.GetCurrency(ctx, userID, inventorydomain.CurrencyType(typeID))
}

// AddCurrencyBalance delegates to inventory activity-point adjustment.
func (s *inventorySpender) AddCurrencyBalance(ctx context.Context, userID int, typeID int, amount int) (int, error) {
	return s.svc.AddCurrencyTracked(ctx, userID, inventorydomain.CurrencyType(typeID), amount, inventorydomain.SourcePurchase, "catalog", "")
}
