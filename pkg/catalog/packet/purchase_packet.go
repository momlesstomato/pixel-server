package packet

import "github.com/momlesstomato/pixel-server/core/codec"

// PurchaseOKPacket defines catalog.purchase_ok (s2c 869) payload.
// It echoes the purchased offer entry back to the client.
type PurchaseOKPacket struct {
	// Offer stores the purchased offer.
	Offer OfferEntry
}

// PacketID returns protocol packet identifier.
func (p PurchaseOKPacket) PacketID() uint16 { return PurchaseOKPacketID }

// Encode serializes purchased offer confirmation into packet body.
func (p PurchaseOKPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := encodeOfferEntry(w, p.Offer); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// PurchaseErrorCode defines catalog purchase error classification.
type PurchaseErrorCode int32

const (
	// PurchaseErrorGeneric defines a generic server-side failure.
	PurchaseErrorGeneric PurchaseErrorCode = 0
	// PurchaseErrorInsufficientCredits defines a credit balance failure.
	PurchaseErrorInsufficientCredits PurchaseErrorCode = 1
	// PurchaseErrorInsufficientPoints defines an activity-point balance failure.
	PurchaseErrorInsufficientPoints PurchaseErrorCode = 2
	// PurchaseErrorNotAvailable defines an offer-not-available failure.
	PurchaseErrorNotAvailable PurchaseErrorCode = 3
)

// PurchaseErrorPacket defines catalog.purchase_error (s2c 1404) payload.
type PurchaseErrorPacket struct {
	// Code stores the purchase error classification.
	Code PurchaseErrorCode
}

// PacketID returns protocol packet identifier.
func (p PurchaseErrorPacket) PacketID() uint16 { return PurchaseErrorPacketID }

// Encode serializes purchase error code into packet body.
func (p PurchaseErrorPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(p.Code))
	return w.Bytes(), nil
}

// PurchaseNotAllowedPacket defines catalog.purchase_not_allowed (s2c 3770) payload.
type PurchaseNotAllowedPacket struct {
	// Code stores the rejection reason code.
	Code int32
}

// PacketID returns protocol packet identifier.
func (p PurchaseNotAllowedPacket) PacketID() uint16 { return PurchaseNotAllowedPacketID }

// Encode serializes not-allowed error code into packet body.
func (p PurchaseNotAllowedPacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(p.Code)
	return w.Bytes(), nil
}
