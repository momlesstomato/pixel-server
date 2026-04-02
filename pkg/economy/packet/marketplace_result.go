package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// MarketplaceBuyResultPacket defines marketplace.buy_result (s2c 2032).
type MarketplaceBuyResultPacket struct {
	// Result stores the buy operation result code. 1 = success, 2 = sold out, 3 = too expensive, 4 = error.
	Result int32
	// OfferID stores the purchased offer identifier.
	OfferID int32
	// NewPrice stores the final price paid.
	NewPrice int32
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceBuyResultPacket) PacketID() uint16 { return MarketplaceBuyResultPacketID }

// Encode serializes buy result into packet body.
func (p MarketplaceBuyResultPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Result)
	w.WriteInt32(p.OfferID)
	w.WriteInt32(p.NewPrice)
	w.WriteInt32(0)
	return w.Bytes(), nil
}

// MarketplaceItemStatsPacket defines marketplace.item_stats (s2c 725).
type MarketplaceItemStatsPacket struct {
	// AvgPrice stores the overall average price.
	AvgPrice int32
	// OfferCount stores the number of active offers.
	OfferCount int32
	// HistoryLength stores the number of history entries.
	HistoryLength int32
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceItemStatsPacket) PacketID() uint16 { return MarketplaceItemStatsPacketID }

// Encode serializes item statistics into packet body.
func (p MarketplaceItemStatsPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.AvgPrice)
	w.WriteInt32(p.OfferCount)
	w.WriteInt32(p.HistoryLength)
	return w.Bytes(), nil
}

// MarketplaceItemPostedPacket defines marketplace.item_posted (s2c 1359).
type MarketplaceItemPostedPacket struct {
	// Result stores the sell result code (1 = success, 2 = error, 3 = limit, 4 = not allowed).
	Result int32
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceItemPostedPacket) PacketID() uint16 { return MarketplaceItemPostedPacketID }

// Encode serializes sell result into packet body.
func (p MarketplaceItemPostedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Result)
	return w.Bytes(), nil
}

// MarketplaceCancelResultPacket defines marketplace.cancel_sale_result (s2c 3264).
type MarketplaceCancelResultPacket struct {
	// OfferID stores the cancelled offer identifier.
	OfferID int32
	// Success stores whether the cancellation succeeded.
	Success bool
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceCancelResultPacket) PacketID() uint16 { return MarketplaceCancelResultPacketID }

// Encode serializes cancel result into packet body.
func (p MarketplaceCancelResultPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.OfferID)
	w.WriteBool(p.Success)
	return w.Bytes(), nil
}

// MarketplaceCanSellPacket defines marketplace.can_sell (s2c 54).
type MarketplaceCanSellPacket struct {
	// ErrorCode stores 1 (can sell), 2 (cannot sell), or 3 (not logged in).
	ErrorCode int32
}

// PacketID returns the wire protocol packet identifier.
func (p MarketplaceCanSellPacket) PacketID() uint16 { return MarketplaceCanSellPacketID }

// Encode serializes can-sell result into packet body.
func (p MarketplaceCanSellPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.ErrorCode)
	w.WriteInt32(p.ErrorCode)
	return w.Bytes(), nil
}
