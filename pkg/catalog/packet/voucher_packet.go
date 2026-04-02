package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// VoucherRedeemOKPacket defines catalog.voucher_redeem_ok (s2c 3336) payload.
type VoucherRedeemOKPacket struct {
	// ProductName stores confirmed voucher product key.
	ProductName string
	// IsHC stores whether voucher reward is HC.
	IsHC bool
}

// PacketID returns protocol packet identifier.
func (p VoucherRedeemOKPacket) PacketID() uint16 { return VoucherRedeemOKPacketID }

// Encode serializes voucher confirmation into packet body.
func (p VoucherRedeemOKPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ProductName); err != nil {
		return nil, err
	}
	if err := w.WriteString(p.ProductName); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// VoucherRedeemErrorPacket defines catalog.voucher_redeem_error (s2c 714) payload.
type VoucherRedeemErrorPacket struct {
	// ErrorCode stores the redemption error classification.
	ErrorCode string
}

// PacketID returns protocol packet identifier.
func (p VoucherRedeemErrorPacket) PacketID() uint16 { return VoucherRedeemErrorPacketID }

// Encode serializes voucher error code into packet body.
func (p VoucherRedeemErrorPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(p.ErrorCode); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// IsOfferGiftablePacket defines catalog.is_offer_giftable (s2c 761) payload.
type IsOfferGiftablePacket struct {
	// OfferID stores the checked catalog offer identifier.
	OfferID int32
	// Giftable stores whether the offer can be gifted.
	Giftable bool
}

// PacketID returns protocol packet identifier.
func (p IsOfferGiftablePacket) PacketID() uint16 { return IsOfferGiftablePacketID }

// Encode serializes giftable check result into packet body.
func (p IsOfferGiftablePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.OfferID)
	w.WriteBool(p.Giftable)
	return w.Bytes(), nil
}
