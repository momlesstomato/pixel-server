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
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
	authenticationhttpapi "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/httpapi"
	authenticationapplication "github.com/momlesstomato/pixel-server/pkg/authentication/application"
	authenticationredisstore "github.com/momlesstomato/pixel-server/pkg/authentication/infrastructure/redisstore"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	"github.com/momlesstomato/pixel-server/pkg/handshake/application/authflow"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
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
	ssoStore, err := authenticationredisstore.NewRedisStore(runtime.Redis, runtime.Config.Authentication.KeyPrefix)
	if err != nil {
		return err
	}
	ssoService := authenticationapplication.NewService(ssoStore, runtime.Config.Authentication)
	if err := registerServeWebSocket(runtime.HTTP, options.WebSocketPath, runtime, ssoService, runtime.Logger); err != nil {
		return err
	}
	if err := authenticationhttpapi.RegisterRoutes(runtime.HTTP, ssoService); err != nil {
		return err
	}
	if err := httpopenapi.RegisterRoutes(runtime.HTTP, httpopenapi.BuildDocument(options.WebSocketPath, authenticationhttpapi.OpenAPIPaths()), "", ""); err != nil {
		return err
	}
	runtime.Logger.Info("http server starting", zap.String("address", fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port)))
	return runServeLifecycle(runtime, runtime.HTTP, fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port), listen)
}

// registerServeWebSocket registers websocket endpoint behavior for serve runtime.
func registerServeWebSocket(module *corehttp.Module, path string, runtime *initializer.Runtime, validator authflow.TicketValidator, logger *zap.Logger) error {
	registry, err := coreconnection.NewRedisSessionRegistryWithOptions(runtime.Redis, coreconnection.RedisSessionRegistryOptions{InstanceID: runtime.Config.App.Name})
	if err != nil {
		return err
	}
	bus, err := handshakerealtime.NewRedisCloseSignalBus(runtime.Redis, "handshake:close")
	if err != nil {
		return err
	}
	handler, err := handshakerealtime.NewHandler(validator, registry, packetsecurity.NewMachineIDPolicy(nil), bus, logger, 30*time.Second)
	if err != nil {
		return err
	}
	broadcaster, err := broadcast.NewRedisBroadcaster(runtime.Redis, "")
	if err != nil {
		return err
	}
	handler.ConfigureBroadcaster(broadcaster)
	statusStore, err := statusredisstore.NewStore(runtime.Redis, runtime.Config.Status.RedisKey)
	if err != nil {
		return err
	}
	hotelStatus, err := sessionhotelstatus.NewService(statusStore, broadcaster, runtime.Config.Status)
	if err != nil {
		return err
	}
	if _, err = hotelStatus.Current(context.Background()); err != nil {
		return err
	}
	hotelStatus.StartCountdownTicker(context.Background())
	userRepository, err := userstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return err
	}
	users, err := userapplication.NewService(userRepository)
	if err != nil {
		return err
	}
	handler.ConfigurePostAuth(hotelStatus, users, runtime.Config.App.Name)
	if webSocketPath := strings.TrimSpace(path); webSocketPath != "" {
		return module.RegisterWebSocket(webSocketPath, handler.Handle)
	}
	return module.RegisterWebSocket("/ws", handler.Handle)
}
