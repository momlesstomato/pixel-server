package packet

import (
	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// MarketplaceConfigPacket defines marketplace.config (s2c 1823) payload.
type MarketplaceConfigPacket struct {
	// Enabled stores whether the marketplace is enabled.
	Enabled bool
	// Commission stores the commission percentage (1 = 1%).
	Commission int32
	// TokenTax stores the token tax rate.
	TokenTax int32
	// OfferMinPrice stores the minimum listing price.
	OfferMinPrice int32
	// OfferMaxPrice stores the maximum listing price.
	OfferMaxPrice int32
	// OfferExpireHours stores the listing expiration hours.
	OfferExpireHours int32
	// AverageDays stores days of price history to display.
	AverageDays int32
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceConfigPacket) PacketID() uint16 { return MarketplaceConfigResponsePacketID }

// Encode serializes marketplace configuration into packet body.
func (p MarketplaceConfigPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteBool(p.Enabled)
	w.WriteInt32(p.Commission)
	w.WriteInt32(p.TokenTax)
	w.WriteInt32(p.OfferMinPrice)
	w.WriteInt32(p.OfferMaxPrice)
	w.WriteInt32(p.OfferExpireHours)
	w.WriteInt32(p.AverageDays)
	return w.Bytes(), nil
}

// MarketplaceSearchResultsPacket defines marketplace.items_searched (s2c 680).
type MarketplaceSearchResultsPacket struct {
	// Offers stores the matching offers.
	Offers []domain.MarketplaceOffer
	// TotalResults stores the total matching offer count.
	TotalResults int
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceSearchResultsPacket) PacketID() uint16 { return MarketplaceSearchResultsPacketID }

// Encode serializes marketplace search results into packet body.
func (p MarketplaceSearchResultsPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Offers)))
	for _, o := range p.Offers {
		encodeMarketplaceOffer(w, o)
	}
	w.WriteInt32(int32(p.TotalResults))
	return w.Bytes(), nil
}

// MarketplaceOwnItemsPacket defines marketplace.own_items (s2c 3884).
type MarketplaceOwnItemsPacket struct {
	// Offers stores the seller's active and completed offers.
	Offers []domain.MarketplaceOffer
	// CreditsWaiting stores uncollected credits from sales.
	CreditsWaiting int
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceOwnItemsPacket) PacketID() uint16 { return MarketplaceOwnItemsPacketID }

// Encode serializes own marketplace items into packet body.
func (p MarketplaceOwnItemsPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.CreditsWaiting))
	w.WriteInt32(int32(len(p.Offers)))
	for _, o := range p.Offers {
		encodeMarketplaceOffer(w, o)
	}
	return w.Bytes(), nil
}

// encodeMarketplaceOffer writes one offer to the writer.
func encodeMarketplaceOffer(w *codec.Writer, o domain.MarketplaceOffer) {
	w.WriteInt32(int32(o.ID))
	w.WriteInt32(1)
	w.WriteInt32(1)
	w.WriteInt32(int32(o.DefinitionID))
	w.WriteInt32(256)
	w.WriteInt32(0)
	w.WriteInt32(int32(o.AskingPrice))
	stateCode := int32(1)
	if o.State == domain.OfferStateSold {
		stateCode = 2
	}
	w.WriteInt32(stateCode)
	w.WriteInt32(int32(o.ExpireAt.Unix() - o.CreatedAt.Unix()))
	w.WriteInt32(int32(o.DefinitionID))
}
