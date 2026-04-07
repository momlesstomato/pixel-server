package cli

import (
	"context"
	"time"

	"github.com/momlesstomato/pixel-server/core/broadcast"
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
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furniturestore "github.com/momlesstomato/pixel-server/pkg/furniture/infrastructure/store"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	handshakerealtime "github.com/momlesstomato/pixel-server/pkg/handshake/adapter/realtime"
	inventoryrealtime "github.com/momlesstomato/pixel-server/pkg/inventory/adapter/realtime"
	inventoryapplication "github.com/momlesstomato/pixel-server/pkg/inventory/application"
	inventorystore "github.com/momlesstomato/pixel-server/pkg/inventory/infrastructure/store"
	navigatorrealtime "github.com/momlesstomato/pixel-server/pkg/navigator/adapter/realtime"
	navigatorapplication "github.com/momlesstomato/pixel-server/pkg/navigator/application"
	navigatorstore "github.com/momlesstomato/pixel-server/pkg/navigator/infrastructure/store"
	subscriptionrealtime "github.com/momlesstomato/pixel-server/pkg/subscription/adapter/realtime"
	subscriptionapplication "github.com/momlesstomato/pixel-server/pkg/subscription/application"
	subscriptionstore "github.com/momlesstomato/pixel-server/pkg/subscription/infrastructure/store"
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
	navigator    *navigatorapplication.Service
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
	catalog.SetItemDeliverer(&furnitureItemDeliverer{svc: furniture})
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
	subscription.SetCreditSpender(&inventorySpender{svc: inventory})
	subscription.SetItemDeliverer(&furnitureItemDeliverer{svc: furniture})
	catalog.SetPurchaseObserver(func(ctx context.Context, userID int, offer catalogdomain.CatalogOffer, amount int) error {
		return subscription.TrackCatalogSpend(ctx, userID, offer.CostCredits*amount)
	})
	navigatorRepo, err := navigatorstore.NewRepository(runtime.PostgreSQL)
	if err != nil {
		return nil, err
	}
	navigator, err := navigatorapplication.NewService(navigatorRepo)
	if err != nil {
		return nil, err
	}
	return &economyServiceBundle{
		furniture: furniture, inventory: inventory, catalog: catalog,
		economy: economy, subscription: subscription, navigator: navigator,
	}, nil
}

// buildEconomyRuntimes creates economy-realm realtime runtimes for packet dispatch.
func buildEconomyRuntimes(bundle *economyServiceBundle, sessions coreconnection.SessionRegistry, transport *handshakerealtime.Transport, broadcaster broadcast.Broadcaster, logger *zap.Logger, liveRoomCount func(int) int) (*furniturerealtime.Runtime, []handshakerealtime.UserRuntime, error) {
	frt, err := furniturerealtime.NewRuntime(bundle.furniture, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	frt.SetBroadcaster(broadcaster)
	irt, err := inventoryrealtime.NewRuntime(bundle.inventory, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	crt, err := catalogrealtime.NewRuntime(bundle.catalog, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	crt.SetInventoryItemSender(func(ctx context.Context, connID string, userID int, itemID int) error {
		item, err := bundle.furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return err
		}
		if item.UserID != userID || item.RoomID != 0 {
			return nil
		}
		def, err := bundle.furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if err != nil {
			return err
		}
		body, err := furnipacket.InventoryAddPacket{
			ItemID: item.ID, SpriteID: def.SpriteID, ExtraData: item.ExtraData,
			AllowRecycle: def.AllowRecycle, AllowTrade: def.AllowTrade,
			AllowInventoryStack:  def.AllowInventoryStack,
			AllowMarketplaceSell: def.AllowMarketplaceSell,
		}.Encode()
		if err != nil {
			return err
		}
		return transport.Send(connID, furnipacket.InventoryAddPacketID, body)
	})
	ert, err := economyrealtime.NewRuntime(bundle.economy, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	srt, err := subscriptionrealtime.NewRuntime(bundle.subscription, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	srt.SetInventoryItemSender(func(ctx context.Context, connID string, userID int, itemID int) error {
		item, err := bundle.furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return err
		}
		if item.UserID != userID || item.RoomID != 0 {
			return nil
		}
		def, err := bundle.furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if err != nil {
			return err
		}
		body, err := furnipacket.InventoryAddPacket{
			ItemID: item.ID, SpriteID: def.SpriteID, ExtraData: item.ExtraData,
			AllowRecycle: def.AllowRecycle, AllowTrade: def.AllowTrade,
			AllowInventoryStack:  def.AllowInventoryStack,
			AllowMarketplaceSell: def.AllowMarketplaceSell,
		}.Encode()
		if err != nil {
			return err
		}
		return transport.Send(connID, furnipacket.InventoryAddPacketID, body)
	})
	nrt, err := navigatorrealtime.NewRuntime(bundle.navigator, sessions, transport, logger)
	if err != nil {
		return nil, nil, err
	}
	nrt.SetLiveRoomCountProvider(liveRoomCount)
	crt.SetClubOffersSender(func(ctx context.Context, connID string, _ int) error {
		return srt.SendClubOffers(ctx, connID)
	})
	return frt, []handshakerealtime.UserRuntime{frt, irt, crt, ert, srt, nrt}, nil
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

// furnitureItemDeliverer adapts furniture.Service to catalogdomain.ItemDeliverer.
type furnitureItemDeliverer struct {
	// svc stores furniture application service.
	svc *furnitureapplication.Service
}

// DeliverItem creates one furniture item instance in the user's inventory.
func (d *furnitureItemDeliverer) DeliverItem(ctx context.Context, userID int, defID int, extraData string, limitedNumber int, limitedTotal int) (int, error) {
	item, err := d.svc.CreateItem(ctx, furnituredomain.Item{
		UserID:        userID,
		DefinitionID:  defID,
		ExtraData:     extraData,
		LimitedNumber: limitedNumber,
		LimitedTotal:  limitedTotal,
	})
	if err != nil {
		return 0, err
	}
	return item.ID, nil
}
