package cli

import (
	"fmt"
	"io"

	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"github.com/momlesstomato/pixel-server/core/logging"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
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
	command := &cobra.Command{Use: "serve", Short: "Start API and websocket server", RunE: func(_ *cobra.Command, _ []string) error {
		options.Output = dependencies.Output
		return ExecuteServe(options, dependencies.Listen)
	}}
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
		Permissions: svc.permissions, EmitPermissionChecked: runtime.Config.Permission.EmitPermissionChecked,
	}}
	pluginManager, err := pluginStage.Initialize()
	if err != nil {
		return err
	}
	defer pluginManager.Shutdown()
	fire := pluginManager.Dispatcher().Fire
	svc.fire = fire
	svc.handler.ConfigurePluginEvents(fire)
	svc.users.SetEventFirer(fire)
	svc.permissions.SetEventFirer(fire)
	svc.sso.SetEventFirer(fire)
	svc.messenger.SetEventFirer(fire)
	svc.furniture.SetEventFirer(fire)
	svc.inventory.SetEventFirer(fire)
	svc.catalog.SetEventFirer(fire)
	svc.economy.SetEventFirer(fire)
	address := fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port)
	runtime.Logger.Info("http server starting", zap.String("address", address))
	return runServeLifecycle(runtime, runtime.HTTP, address, listen)
}
