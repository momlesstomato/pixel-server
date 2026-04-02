package realtime

import (
	"strings"

	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
	"github.com/momlesstomato/pixel-server/pkg/catalog/packet"
)

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
		PriceCredits: int32(o.CostCredits), PriceActivityPoints: int32(o.CostActivityPoints),
		ActivityPointType: int32(o.ActivityPointType), Giftable: true,
		Products: products, CanSelectAmount: o.Amount > 1,
	}
}

// buildOfferProducts builds the product list for one offer.
func buildOfferProducts(o domain.CatalogOffer) []packet.OfferProduct {
	if o.BadgeID != "" {
		return []packet.OfferProduct{{TypeCode: "b", IsBadge: true, BadgeCode: o.BadgeID}}
	}
	typeCode := o.ItemType
	if typeCode == "" {
		typeCode = "s"
	}
	return []packet.OfferProduct{{
		TypeCode: typeCode, SpriteID: int32(o.SpriteID),
		ExtraData: o.ExtraData, Amount: int32(o.Amount),
		IsLimited: o.IsLimited(), LimitedTotal: int32(o.LimitedTotal),
		LimitedRemaining: int32(o.LimitedTotal - o.LimitedSells),
	}}
}
