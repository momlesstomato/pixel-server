package cli

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/core/initializer"
	authenticationapplication "github.com/momlesstomato/pixel-server/pkg/authentication/application"
	authenticationredisstore "github.com/momlesstomato/pixel-server/pkg/authentication/infrastructure/redisstore"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	permissionnotification "github.com/momlesstomato/pixel-server/pkg/permission/adapter/notification"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissionstore "github.com/momlesstomato/pixel-server/pkg/permission/infrastructure/store"
	sessionhotelstatus "github.com/momlesstomato/pixel-server/pkg/status/application/hotelstatus"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
)

// serveServices holds shared dependencies built during serve startup.
type serveServices struct {
	sso         *authenticationapplication.Service
	registry    *coreconnection.RedisSessionRegistry
	bus         *handshakerealtime.DistributedCloseSignalBus
	broadcaster broadcast.Broadcaster
	hotelStatus *sessionhotelstatus.Service
	users       *userapplication.Service
	permissions *permissionapplication.Service
	handler     *handshakerealtime.Handler
}

// FindByID verifies user existence by identifier, satisfying the authflow.UserFinder interface.
func (services *serveServices) FindByID(ctx context.Context, id int) error {
	_, err := services.users.FindByID(ctx, id)
	return err
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
		CachePrefix:          runtime.Config.Permission.CachePrefix,
		CacheTTL:             time.Duration(runtime.Config.Permission.CacheTTLSeconds) * time.Second,
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
	return &serveServices{
		sso:      authenticationapplication.NewService(ssoStore, runtime.Config.Authentication),
		registry: registry, bus: bus, broadcaster: broadcaster,
		hotelStatus: hotelStatus, users: users, permissions: permissions,
	}, nil
}
