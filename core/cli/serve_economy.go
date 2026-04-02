package cli

import (
	"context"
	"time"

	coreconnection "github.com/momlesstomato/pixel-server/core/connection"
	"github.com/momlesstomato/pixel-server/core/initializer"
	catalogrealtime "github.com/momlesstomato/pixel-server/pkg/catalog/adapter/realtime"
	catalogapplication "github.com/momlesstomato/pixel-server/pkg/catalog/application"
	catalogdomain "github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	catalogstore "github.com/momlesstomato/pixel-server/pkg/catalog/infrastructure/store"
	economyrealtime "github.com/momlesstomato/pixel-server/pkg/economy/adapter/realtime"
	economyapplication "github.com/momlesstomato/pixel-server/pkg/economy/application"
	economystore "github.com/momlesstomato/pixel-server/pkg/economy/infrastructure/store"
	furniturerealtime "github.com/momlesstomato/pixel-server/pkg/furniture/adapter/realtime"
	furnitureapplication "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furniturestore "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/store"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	inventoryrealtime "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/realtime"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	inventorydomain "github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	inventorystore "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/store"
	subscriptionrealtime "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/realtime"
	subscriptionapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	subscriptionstore "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/store"
	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	userstore "github.com/momlesstomato/pixel-server/pkg/user/infrastructure/store"
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
	catalog.SetCache(runtime.Redis, catalogapplication.CacheConfig{Prefix: "catalog", TTL: 5 * time.Minute})
	catalog.SetCurrencyValidator(inventoryRepo)
	catalog.SetSpender(&inventorySpender{svc: inventory})
	userRepo, err := userstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	catalog.SetRecipientFinder(&userRecipientFinder{repo: userRepo})
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

