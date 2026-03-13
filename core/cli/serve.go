package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
	"github.com/momlesstomato/pixel-server/core/config"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	httpopenapi "github.com/momlesstomato/pixel-server/core/http/openapi"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"github.com/momlesstomato/pixel-server/core/logging"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
	authenticationhttpapi "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/httpapi"
	authenticationapplication "github.com/momlesstomato/pixel-server/pkg/authentication/application"
	authenticationredisstore "github.com/momlesstomato/pixel-server/pkg/authentication/infrastructure/redisstore"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	managementhttpapi "github.com/momlesstomato/pixel-server/pkg/management/adapter/httpapi"
	sessionhotelstatus "github.com/momlesstomato/pixel-server/pkg/status/application/hotelstatus"
	statusredisstore "github.com/momlesstomato/pixel-server/pkg/status/infrastructure/redisstore"
	userapplication "github.com/momlesstomato/pixel-server/pkg/user/application"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// ServeListenFunc defines network startup behavior for the HTTP module.
type ServeListenFunc func(*corehttp.Module, string) error

// ServeDependencies defines runtime overrides for the serve command.
type ServeDependencies struct {
	// Listen overrides Fiber listen behavior for integration and tests.
	Listen ServeListenFunc
	// Output overrides logger output stream.
	Output io.Writer
}

// ServeOptions defines startup inputs for serve execution.
type ServeOptions struct {
	// EnvFile defines the config file path.
	EnvFile string
	// EnvPrefix defines optional environment prefix.
	EnvPrefix string
	// WebSocketPath defines websocket endpoint route.
	WebSocketPath string
	// APIKeyHeader defines the header name used for API key transport.
	APIKeyHeader string
	// Output defines logger output stream.
	Output io.Writer
}

// NewServeCommand creates the serve subcommand.
func NewServeCommand(dependencies ServeDependencies) *cobra.Command {
	var options ServeOptions
	command := &cobra.Command{
		Use:   "serve",
		Short: "Start API and websocket server",
		RunE: func(_ *cobra.Command, _ []string) error {
			options.Output = dependencies.Output
			return ExecuteServe(options, dependencies.Listen)
		},
	}
	command.Flags().StringVar(&options.EnvFile, "env-file", ".env", "Environment file path")
	command.Flags().StringVar(&options.EnvPrefix, "env-prefix", "", "Environment key prefix")
	command.Flags().StringVar(&options.WebSocketPath, "ws-path", "/ws", "WebSocket endpoint path")
	command.Flags().StringVar(&options.APIKeyHeader, "api-key-header", corehttp.DefaultAPIKeyHeader, "API key header name")
	return command
}

// ExecuteServe initializes dependencies and starts the Fiber server.
func ExecuteServe(options ServeOptions, listen ServeListenFunc) error {
	runner := initializer.NewRunner(
		config.Initializer{Options: config.LoaderOptions{EnvFile: options.EnvFile, EnvPrefix: options.EnvPrefix}},
		rediscore.Initializer{},
		logging.Initializer{Output: options.Output},
		postgrescore.Initializer{},
		corehttp.Initializer{APIKeyHeader: options.APIKeyHeader},
	)
	runtime, err := runner.Run()
	if err != nil {
		return err
	}
	svc, err := buildServeServices(runtime)
	if err != nil {
		return err
	}
	if err := registerServeWebSocket(runtime.HTTP, options.WebSocketPath, runtime, svc); err != nil {
		return err
	}
	if err := registerServeHTTPRoutes(runtime.HTTP, svc, options.WebSocketPath); err != nil {
		return err
	}
	pluginStage := coreplugin.Stage{Dir: "plugins", Logger: runtime.Logger, Deps: coreplugin.ServerDependencies{
		Registry: svc.registry, Broadcaster: svc.broadcaster, BroadcastChannel: runtime.Config.Status.BroadcastChannel,
	}}
	pluginManager, err := pluginStage.Initialize()
	if err != nil {
		return err
	}
	defer pluginManager.Shutdown()
	addr := fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port)
	runtime.Logger.Info("http server starting", zap.String("address", addr))
	return runServeLifecycle(runtime, runtime.HTTP, addr, listen)
}

// serveServices holds shared dependencies built during serve startup.
type serveServices struct {
	sso         *authenticationapplication.Service
	registry    *coreconnection.RedisSessionRegistry
	bus         *handshakerealtime.DistributedCloseSignalBus
	broadcaster broadcast.Broadcaster
	hotelStatus *sessionhotelstatus.Service
	users       *userapplication.Service
}

// buildServeServices constructs shared application dependencies.
func buildServeServices(rt *initializer.Runtime) (*serveServices, error) {
	ssoStore, err := authenticationredisstore.NewRedisStore(rt.Redis, rt.Config.Authentication.KeyPrefix)
	if err != nil {
		return nil, err
	}
	registry, err := coreconnection.NewRedisSessionRegistryWithOptions(rt.Redis, coreconnection.RedisSessionRegistryOptions{InstanceID: rt.Config.App.Name})
	if err != nil {
		return nil, err
	}
	bus, err := handshakerealtime.NewRedisCloseSignalBus(rt.Redis, "handshake:close")
	if err != nil {
		return nil, err
	}
	broadcaster, err := broadcast.NewRedisBroadcaster(rt.Redis, "")
	if err != nil {
		return nil, err
	}
	statusStore, err := statusredisstore.NewStore(rt.Redis, rt.Config.Status.RedisKey)
	if err != nil {
		return nil, err
	}
	hotel, err := sessionhotelstatus.NewService(statusStore, broadcaster, rt.Config.Status)
	if err != nil {
		return nil, err
	}
	if _, err = hotel.Current(context.Background()); err != nil {
		return nil, err
	}
	hotel.StartCountdownTicker(context.Background())
	userRepo, err := userstore.NewRepository(rt.PostgreSQL)
	if err != nil {
		return nil, err
	}
	users, err := userapplication.NewService(userRepo)
	if err != nil {
		return nil, err
	}
	return &serveServices{
		sso:      authenticationapplication.NewService(ssoStore, rt.Config.Authentication),
		registry: registry, bus: bus, broadcaster: broadcaster,
		hotelStatus: hotel, users: users,
	}, nil
}

// registerServeWebSocket registers websocket endpoint behavior.
func registerServeWebSocket(module *corehttp.Module, path string, rt *initializer.Runtime, svc *serveServices) error {
	handler, err := handshakerealtime.NewHandler(svc.sso, svc.registry, packetsecurity.NewMachineIDPolicy(nil), svc.bus, rt.Logger, 30*time.Second)
	if err != nil {
		return err
	}
	handler.ConfigureBroadcaster(svc.broadcaster)
	handler.ConfigurePostAuth(svc.hotelStatus, svc.users, rt.Config.App.Name)
	wsPath := strings.TrimSpace(path)
	if wsPath == "" {
		wsPath = "/ws"
	}
	return module.RegisterWebSocket(wsPath, handler.Handle)
}

// registerServeHTTPRoutes registers all REST API routes and OpenAPI documentation.
func registerServeHTTPRoutes(module *corehttp.Module, svc *serveServices, wsPath string) error {
	if err := authenticationhttpapi.RegisterRoutes(module, svc.sso); err != nil {
		return err
	}
	closer := &busCloserAdapter{bus: svc.bus}
	if err := managementhttpapi.RegisterSessionRoutes(module, svc.registry, closer); err != nil {
		return err
	}
	if err := managementhttpapi.RegisterHotelRoutes(module, svc.hotelStatus); err != nil {
		return err
	}
	paths := mergeOpenAPIPaths(authenticationhttpapi.OpenAPIPaths(), managementhttpapi.OpenAPIPaths())
	return httpopenapi.RegisterRoutes(module, httpopenapi.BuildDocument(wsPath, paths), "", "")
}

// busCloserAdapter adapts DistributedCloseSignalBus to SessionCloser interface.
type busCloserAdapter struct {
	bus *handshakerealtime.DistributedCloseSignalBus
}

// Close publishes a close signal for one connection identifier.
func (a *busCloserAdapter) Close(ctx context.Context, connID string, code int, reason string) error {
	return a.bus.Publish(ctx, connID, handshakerealtime.CloseSignal{Code: code, Reason: reason})
}

// mergeOpenAPIPaths combines multiple OpenAPI path maps.
func mergeOpenAPIPaths(maps ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}
