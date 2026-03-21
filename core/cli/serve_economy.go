package cli

import (
	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/core/initializer"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	catalogrealtime "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/realtime"
	catalogstore "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/store"
	economyapplication "github.com/momlesstomato/pixel-server/pkg/economy/application"
	economyrealtime "github.com/momlesstomato/pixel-server/pkg/economy/adapter/realtime"
	economystore "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/store"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furniturerealtime "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furniturestore "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/store"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	inventoryrealtime "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/realtime"
	inventorystore "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/store"
	subscriptionapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	subscriptionrealtime "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/realtime"
	subscriptionstore "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/store"
	"go.uber.org/zap"
)

// economyServiceBundle groups economy-realm application services.
type economyServiceBundle struct {
	furniture    *furnitureapplication.Service
	inventory    *inventoryapplication.Service
	catalog      *catalogapplication.Service
	economy      *economyapplication.Service
	subscription *subscriptionapplication.Service
}

// buildEconomyServices constructs economy-realm application services.
func buildEconomyServices(runtime *initializer.Runtime) (*economyServiceBundle, error) {
	furnitureRepo, err := furniturestore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	furniture, err := furnitureapplication.NewService(furnitureRepo)
	if err != nil {
		return nil, err
	}
	inventoryRepo, err := inventorystore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	inventory, err := inventoryapplication.NewService(inventoryRepo)
	if err != nil {
		return nil, err
	}
	catalogRepo, err := catalogstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	catalog, err := catalogapplication.NewService(catalogRepo)
	if err != nil {
		return nil, err
	}
	economyRepo, err := economystore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	economy, err := economyapplication.NewService(economyRepo)
	if err != nil {
		return nil, err
	}
	subscriptionRepo, err := subscriptionstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	subscription, err := subscriptionapplication.NewService(subscriptionRepo)
	if err != nil {
		return nil, err
	}
	return &economyServiceBundle{
		furniture: furniture, inventory: inventory, catalog: catalog,
		economy: economy, subscription: subscription,
	}, nil
}

// mergeOpenAPIPaths combines multiple OpenAPI path maps.
func mergeOpenAPIPaths(maps ...map[string]any) map[string]any {
	merged := map[string]any{}
	for _, value := range maps {
		for path, pathItem := range value {
			merged[path] = pathItem
		}
	}
	return merged
}

// buildEconomyRuntimes creates economy-realm realtime runtimes for packet dispatch.
func buildEconomyRuntimes(bundle *economyServiceBundle, sessions coreconnection.SessionRegistry, transport *handshakerealtime.Transport, logger *zap.Logger) ([]handshakerealtime.UserRuntime, error) {
	frt, err := furniturerealtime.NewRuntime(bundle.furniture, sessions, transport, logger)
	if err != nil {
		return nil, err
	}
	irt, err := inventoryrealtime.NewRuntime(bundle.inventory, sessions, transport, logger)
	if err != nil {
		return nil, err
	}
	crt, err := catalogrealtime.NewRuntime(bundle.catalog, sessions, transport, logger)
	if err != nil {
		return nil, err
	}
	ert, err := economyrealtime.NewRuntime(bundle.economy, sessions, transport, logger)
	if err != nil {
		return nil, err
	}
	srt, err := subscriptionrealtime.NewRuntime(bundle.subscription, sessions, transport, logger)
	if err != nil {
		return nil, err
	}
	return []handshakerealtime.UserRuntime{frt, irt, crt, ert, srt}, nil
}
