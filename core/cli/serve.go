package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/momlesstomato/pixel-server/core/config"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/core/initializer"
	"github.com/momlesstomato/pixel-server/core/logging"
	coreplugin "github.com/momlesstomato/pixel-server/core/plugin"
	postgrescore "github.com/momlesstomato/pixel-server/core/postgres"
	rediscore "github.com/momlesstomato/pixel-server/core/redis"
	authcommand "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/command"
	catalogcommand "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/command"
	catalogdomain "github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	economycommand "github.com/momlesstomato/pixel-server/pkg/economy/adapter/command"
	furniturecommand "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/command"
	inventorycommand "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/command"
	messengercommand "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/command"
	navigatorcommand "github.com/momlesstomato/pixel-server/pkg/navigator/adapter/command"
	permissioncommand "github.com/momlesstomato/pixel-server/pkg/permission/adapter/command"
	subscriptioncommand "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/command"
	usercommand "github.com/momlesstomato/pixel-server/pkg/user/adapter/command"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
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
	svc.navigator.SetEventFirer(fire)
	address := fmt.Sprintf("%s:%d", runtime.Config.App.BindIP, runtime.Config.App.Port)
	runtime.Logger.Info("http server starting", zap.String("address", address))
	return runServeLifecycle(runtime, runtime.HTTP, address, listen)
}

// Dependencies defines root command dependencies.
type Dependencies struct {
	// Serve defines serve command runtime dependencies.
	Serve ServeDependencies
	// Authentication defines SSO command runtime dependencies.
	Authentication authcommand.Dependencies
	// User defines user command runtime dependencies.
	User usercommand.Dependencies
	// Permission defines permission group command runtime dependencies.
	Permission permissioncommand.Dependencies
	// Messenger defines messenger command runtime dependencies.
	Messenger messengercommand.Dependencies
	// Furniture defines furniture command runtime dependencies.
	Furniture furniturecommand.Dependencies
	// Inventory defines inventory command runtime dependencies.
	Inventory inventorycommand.Dependencies
	// Catalog defines catalog command runtime dependencies.
	Catalog catalogcommand.Dependencies
	// Economy defines economy command runtime dependencies.
	Economy economycommand.Dependencies
	// Subscription defines subscription command runtime dependencies.
	Subscription subscriptioncommand.Dependencies
	// Navigator defines navigator command runtime dependencies.
	Navigator navigatorcommand.Dependencies
}

// NewRootCommand creates the root CLI command tree.
func NewRootCommand(dependencies Dependencies) *cobra.Command {
	command := &cobra.Command{
		Use:   "pixel-server",
		Short: "Pixel server runtime CLI",
	}
	command.AddCommand(NewServeCommand(dependencies.Serve))
	command.AddCommand(NewDBCommand())
	command.AddCommand(authcommand.NewSSOCommand(dependencies.Authentication))
	command.AddCommand(usercommand.NewUserCommand(dependencies.User))
	command.AddCommand(permissioncommand.NewGroupCommand(dependencies.Permission))
	command.AddCommand(messengercommand.NewMessengerCommand(dependencies.Messenger))
	command.AddCommand(furniturecommand.NewFurnitureCommand(dependencies.Furniture))
	command.AddCommand(inventorycommand.NewInventoryCommand(dependencies.Inventory))
	command.AddCommand(catalogcommand.NewCatalogCommand(dependencies.Catalog))
	command.AddCommand(economycommand.NewEconomyCommand(dependencies.Economy))
	command.AddCommand(subscriptioncommand.NewSubscriptionCommand(dependencies.Subscription))
	command.AddCommand(navigatorcommand.NewNavigatorCommand(dependencies.Navigator))
	return command
}

// userRecipientFinder adapts user.Repository to catalogdomain.RecipientFinder.
type userRecipientFinder struct {
	// repo stores user repository.
	repo userdomain.Repository
}

// FindRecipientByUsername resolves a catalog recipient by username.
func (f *userRecipientFinder) FindRecipientByUsername(ctx context.Context, username string) (catalogdomain.RecipientInfo, error) {
	user, err := f.repo.FindByUsername(ctx, username)
	if err != nil {
		return catalogdomain.RecipientInfo{}, catalogdomain.ErrRecipientNotFound
	}
	return catalogdomain.RecipientInfo{UserID: user.ID, AllowGifts: !user.SafetyLocked}, nil
}
