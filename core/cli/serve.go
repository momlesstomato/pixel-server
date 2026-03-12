package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/momlesstomato/pixel-server/core/config"
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	httpopenapi "github.com/momlesstomato/pixel-server/core/http/openapi"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"github.com/momlesstomato/pixel-server/core/logging"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
	"github.com/momlesstomato/pixel-server/pkg/authentication"
	"github.com/momlesstomato/pixel-server/pkg/authentication/httpapi"
	"github.com/momlesstomato/pixel-server/pkg/handshake/authflow"
	packetsecurity "github.com/momlesstomato/pixel-server/pkg/handshake/packet/security"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/realtime"
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
		corehttp.Initializer{APIKeyHeader: options.APIKeyHeader},
	)
	runtime, err := runner.Run()
	if err != nil {
		return err
	}
	ssoStore, err := authentication.NewRedisStore(runtime.Redis, runtime.Config.Authentication.KeyPrefix)
	if err != nil {
		return err
	}
	ssoService := authentication.NewService(ssoStore, runtime.Config.Authentication)
	module := runtime.HTTP
	if err := registerServeWebSocket(module, options.WebSocketPath, runtime, ssoService, runtime.Logger); err != nil {
		return err
	}
	if err := httpapi.RegisterRoutes(module, ssoService); err != nil {
		return err
	}
	openAPIDocument := httpopenapi.BuildDocument(options.WebSocketPath, httpapi.OpenAPIPaths())
	if err := httpopenapi.RegisterRoutes(module, openAPIDocument, "", ""); err != nil {
		return err
	}
	address := fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port)
	runtime.Logger.Info("http server starting", zap.String("address", address))
	return runServeLifecycle(runtime, module, address, listen)
}

// EchoWebSocketHandler mirrors inbound messages to the same connection.
func EchoWebSocketHandler(connection *websocket.Conn) {
	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil {
			return
		}
		if err := connection.WriteMessage(messageType, payload); err != nil {
			return
		}
	}
}

// defaultListen starts Fiber on the configured bind address.
func defaultListen(module *corehttp.Module, address string) error {
	return module.App().Listen(address)
}

// registerServeWebSocket registers websocket endpoint behavior for serve runtime.
func registerServeWebSocket(module *corehttp.Module, path string, runtime *initializer.Runtime, validator authflow.TicketValidator, logger *zap.Logger) error {
	registry, err := coreconnection.NewRedisSessionRegistry(runtime.Redis)
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
	webSocketPath := strings.TrimSpace(path)
	if webSocketPath == "" {
		webSocketPath = "/ws"
	}
	return module.RegisterWebSocket(webSocketPath, handler.Handle)
}
