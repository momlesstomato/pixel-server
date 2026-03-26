package realtime

import (
	"context"
	"strings"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
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

// buildIndexTree constructs a root IndexNode from a flat list of pages.
func buildIndexTree(pages []domain.CatalogPage) packet.IndexNode {
	byParent := make(map[int][]domain.CatalogPage)
	for _, p := range pages {
		key := -1
		if p.ParentID != nil {
			key = *p.ParentID
		}
		byParent[key] = append(byParent[key], p)
	}
	roots := byParent[-1]
	children := make([]packet.IndexNode, 0, len(roots))
	for _, p := range roots {
		children = append(children, buildSubTree(p, byParent))
	}
	return packet.IndexNode{Visible: true, Icon: 0, PageID: -1, PageName: "root", Caption: "", Children: children}
}

// buildSubTree recursively converts one page and its children into IndexNodes.
func buildSubTree(page domain.CatalogPage, byParent map[int][]domain.CatalogPage) packet.IndexNode {
	sub := byParent[page.ID]
	children := make([]packet.IndexNode, 0, len(sub))
	for _, child := range sub {
		children = append(children, buildSubTree(child, byParent))
	}
	pageID := int32(page.ID)
	if !page.Enabled {
		pageID = -1
	}
	return packet.IndexNode{
		Visible: page.Visible, Icon: int32(page.IconImage),
		PageID: pageID, PageName: strings.ToLower(page.Caption),
		Caption: page.Caption, Children: children,
	}
}

// buildPagePacket converts domain page and offers into a PagePacket.
func buildPagePacket(page domain.CatalogPage, offers []domain.CatalogOffer, catalogType string) packet.PagePacket {
	entries := make([]packet.OfferEntry, 0, len(offers))
	for _, o := range offers {
		entries = append(entries, buildOfferEntry(o))
	}
	return packet.PagePacket{
		PageID: int32(page.ID), CatalogType: catalogType,
		LayoutCode: page.PageLayout, Images: page.Images,
		Texts: page.Texts, Offers: entries,
	}
}

// buildOfferEntry converts one domain offer into an OfferEntry.
func buildOfferEntry(o domain.CatalogOffer) packet.OfferEntry {
	products := buildOfferProducts(o)
	return packet.OfferEntry{
		OfferID: int32(o.ID), LocalizationID: o.CatalogName,
		PriceCredits: int32(o.CostPrimary), PricePoints: int32(o.CostSecondary),
		PointType: int32(o.CostSecondaryType), Giftable: false,
		Products: products, CanSelectAmount: o.Amount > 1,
	}
}

// buildOfferProducts builds the product list for one offer.
func buildOfferProducts(o domain.CatalogOffer) []packet.OfferProduct {
	if o.BadgeID != "" {
		return []packet.OfferProduct{{TypeCode: "b", IsBadge: true, BadgeCode: o.BadgeID}}
	}
	return []packet.OfferProduct{{
		TypeCode: "i", SpriteID: int32(o.ItemDefinitionID),
		ExtraData: o.ExtraData, Amount: int32(o.Amount),
		IsLimited: o.IsLimited(), LimitedTotal: int32(o.LimitedTotal),
		LimitedRemaining: int32(o.LimitedTotal - o.LimitedSells),
	}}
}
