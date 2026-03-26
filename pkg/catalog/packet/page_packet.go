package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// OfferProduct defines one product entry within a catalog offer.
type OfferProduct struct {
	// TypeCode stores item type character: "i" floor, "s" wall, "b" badge.
	TypeCode string
	// SpriteID stores furniture sprite identifier for non-badge products.
	SpriteID int32
	// ExtraData stores item custom data payload.
	ExtraData string
	// Amount stores number of copies.
	Amount int32
	// IsBadge stores whether this product is a badge.
	IsBadge bool
	// BadgeCode stores badge code when IsBadge is true.
	BadgeCode string
	// IsLimited stores limited edition flag.
	IsLimited bool
	// LimitedTotal stores total limited edition run.
	LimitedTotal int32
	// LimitedRemaining stores remaining limited copies.
	LimitedRemaining int32
}

// OfferEntry defines one offer encoded inside a catalog page response.
type OfferEntry struct {
	// OfferID stores offer identifier.
	OfferID int32
	// LocalizationID stores the offer caption string.
	LocalizationID string
	// PriceCredits stores credit price.
	PriceCredits int32
	// PricePoints stores activity point price.
	PricePoints int32
	// PointType stores activity point currency type.
	PointType int32
	// Giftable stores whether the offer can be gifted.
	Giftable bool
	// Products stores individual item products.
	Products []OfferProduct
	// CanSelectAmount stores whether multi-buy is allowed.
	CanSelectAmount bool
}

// PagePacket defines catalog.page (s2c 804) payload.
type PagePacket struct {
	// PageID stores page identifier.
	PageID int32
	// CatalogType stores echoed catalog mode string.
	CatalogType string
	// LayoutCode stores layout template name.
	LayoutCode string
	// Images stores page image asset names.
	Images []string
	// Texts stores page text blocks.
	Texts []string
	// Offers stores the purchasable offers on this page.
	Offers []OfferEntry
}

// PacketID returns protocol packet identifier.
func (p PagePacket) PacketID() uint16 { return PageResponsePacketID }

// Encode serializes catalog page content into packet body.
func (p PagePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.PageID)
	if err := w.WriteString(p.CatalogType); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.LayoutCode); err != nil {
		return nil, err
	}
	w.WriteInt32(int32(len(p.Images)))
	for _, img := range p.Images {
		if err := w.WriteString(img); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Texts)))
	for _, txt := range p.Texts {
		if err := w.WriteString(txt); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(int32(len(p.Offers)))
	for _, offer := range p.Offers {
		if err := encodeOfferEntry(w, offer); err != nil {
			return nil, err
		}
	}
	w.WriteInt32(-1)
	w.WriteBool(false)
	return w.Bytes(), nil
}

// encodeOfferEntry writes one offer record into the packet writer.
func encodeOfferEntry(w *codec.Writer, offer OfferEntry) error {
	w.WriteInt32(offer.OfferID)
	if err := w.WriteString(offer.LocalizationID); err != nil {
		return err
	}
	w.WriteBool(false)
	w.WriteInt32(offer.PriceCredits)
	w.WriteInt32(offer.PricePoints)
	w.WriteInt32(offer.PointType)
	w.WriteBool(offer.Giftable)
	w.WriteInt32(int32(len(offer.Products)))
	for _, prod := range offer.Products {
		if err := encodeOfferProduct(w, prod); err != nil {
			return err
		}
	}
	w.WriteInt32(0)
	w.WriteBool(offer.CanSelectAmount)
	w.WriteBool(false)
	return w.WriteString("")
}

// encodeOfferProduct writes one product entry into the packet writer.
func encodeOfferProduct(w *codec.Writer, prod OfferProduct) error {
	if err := w.WriteString(prod.TypeCode); err != nil {
		return err
	}
	if prod.IsBadge {
		return w.WriteString(prod.BadgeCode)
	}
	w.WriteInt32(prod.SpriteID)
	if err := w.WriteString(prod.ExtraData); err != nil {
		return err
	}
	w.WriteInt32(prod.Amount)
	w.WriteBool(prod.IsLimited)
	if prod.IsLimited {
		w.WriteInt32(prod.LimitedTotal)
		w.WriteInt32(prod.LimitedRemaining)
	}
	return nil
}
