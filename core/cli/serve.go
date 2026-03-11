package cli

import (
	"fmt"
	"io"

	"github.com/gofiber/contrib/websocket"
	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"github.com/momlesstomato/pixel-server/core/logging"
	"github.com/spf13/cobra"
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
	// APIKey defines the required API key for every route.
	APIKey string
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
	command.Flags().StringVar(&options.APIKey, "api-key", "", "Required API key for all routes")
	command.Flags().StringVar(&options.APIKeyHeader, "api-key-header", corehttp.DefaultAPIKeyHeader, "API key header name")
	_ = command.MarkFlagRequired("api-key")
	return command
}

// ExecuteServe initializes dependencies and starts the Fiber server.
func ExecuteServe(options ServeOptions, listen ServeListenFunc) error {
	runner := initializer.NewRunner(
		config.Initializer{Options: config.LoaderOptions{
			EnvFile: options.EnvFile, EnvPrefix: options.EnvPrefix,
		}},
		logging.Initializer{Output: options.Output},
		corehttp.Initializer{APIKey: options.APIKey, APIKeyHeader: options.APIKeyHeader},
		corehttp.WebSocketInitializer{Path: options.WebSocketPath, Handler: EchoWebSocketHandler},
	)
	runtime, err := runner.Run()
	if err != nil {
		return err
	}
	cfg := runtime.Config
	module := runtime.HTTP
	if listen == nil {
		listen = defaultListen
	}
	address := fmt.Sprintf("%s:%d", cfg.App.BindIP, cfg.App.Port)
	return listen(module, address)
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
