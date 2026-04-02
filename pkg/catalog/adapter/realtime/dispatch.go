package realtime

import (
	"context"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
	"go.uber.org/zap"
)

// Handle dispatches one authenticated catalog packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	userID, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case packet.GetIndexPacketID:
		return true, runtime.handleGetIndex(ctx, connID, userID, body)
	case packet.GetPagePacketID:
		return true, runtime.handleGetPage(ctx, connID, userID, body)
	case packet.GetGiftWrappingConfigPacketID:
		return true, runtime.handleGetGiftWrappingConfig(connID)
	case packet.PurchasePacketID:
		return true, runtime.handlePurchase(ctx, connID, userID, body)
	case packet.PurchaseGiftPacketID:
		return true, runtime.handlePurchaseGift(ctx, connID, userID, body)
	default:
		return false, nil
	}
}

// handleGetIndex responds with the catalog page tree.
func (runtime *Runtime) handleGetIndex(ctx context.Context, connID string, userID int, body []byte) error {
	catalogType := parseCatalogType(body)
	pages, err := runtime.service.ListPages(ctx)
	if err != nil {
		runtime.logger.Error("list catalog pages failed", zap.Int("user_id", userID), zap.Error(err))
		return err
	}
	root := buildIndexTree(pages)
	return runtime.sendPacket(connID, packet.IndexPacket{Root: root, NewItems: false, CatalogType: catalogType})
}

// handleGetPage responds with catalog page content.
func (runtime *Runtime) handleGetPage(ctx context.Context, connID string, userID int, body []byte) error {
	pageID, catalogType, err := parseGetPageRequest(body)
	if err != nil {
		return err
	}
	page, err := runtime.service.FindPageByID(ctx, int(pageID))
	if err != nil {
		runtime.logger.Error("find catalog page failed", zap.Int("user_id", userID), zap.Int32("page_id", pageID), zap.Error(err))
		return err
	}
	offers, err := runtime.service.ListOffersByPageID(ctx, int(pageID))
	if err != nil {
		runtime.logger.Error("list catalog offers failed", zap.Int("user_id", userID), zap.Int32("page_id", pageID), zap.Error(err))
		return err
	}
	return runtime.sendPacket(connID, buildPagePacket(page, offers, catalogType))
}

// parseCatalogType reads the catalog mode string from a get_index body.
func parseCatalogType(body []byte) string {
	reader := codec.NewReader(body)
	s, err := reader.ReadString()
	if err != nil || s == "" {
		return "NORMAL"
	}
	return s
}

// parseGetPageRequest reads pageId and catalogType from a get_page body.
func parseGetPageRequest(body []byte) (int32, string, error) {
	reader := codec.NewReader(body)
	pageID, err := reader.ReadInt32()
	if err != nil {
		return 0, "", err
	}
	_, err = reader.ReadInt32()
	if err != nil {
		return 0, "", err
	}
	catalogType, err := reader.ReadString()
	if err != nil {
		return pageID, "NORMAL", nil
	}
	return pageID, catalogType, nil
}
