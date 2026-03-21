package cli

import (
	authcommand "github.com/momlesstomato/pixel-server/pkg/authentication/adapter/command"
	catalogcommand "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/command"
	economycommand "github.com/momlesstomato/pixel-server/pkg/economy/adapter/command"
	furniturecommand "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/command"
	inventorycommand "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/command"
	messengercommand "github.com/momlesstomato/pixel-server/pkg/messenger/adapter/command"
	permissioncommand "github.com/momlesstomato/pixel-server/pkg/permission/adapter/command"
	subscriptioncommand "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/command"
	usercommand "github.com/momlesstomato/pixel-server/pkg/user/adapter/command"
	"github.com/spf13/cobra"
)

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
	return command
}
