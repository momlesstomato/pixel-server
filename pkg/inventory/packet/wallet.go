package packet

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// CreditBalancePacket encodes the user credit balance for wire protocol packet 3475.
// The client expects a string of the form "{amount}.0".
type CreditBalancePacket struct {
	// Balance stores the user credit balance.
	Balance int
}

// PacketID returns the wire protocol packet identifier.
func (p CreditBalancePacket) PacketID() uint16 { return CreditsResponsePacketID }

// Encode serializes the credit balance as a decimal string with a .0 suffix.
func (p CreditBalancePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	if err := w.WriteString(fmt.Sprintf("%d.0", p.Balance)); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// CurrencyBalancePacket encodes activity-point balances for wire protocol packet 2018.
// Each entry is a pair of (type_id int32, amount int32) preceded by an entry count.
// Credits (type -1) must not be included; they are sent via CreditBalancePacket.
type CurrencyBalancePacket struct {
	// Currencies stores the activity-point currency entries to send.
	Currencies []domain.Currency
}

// PacketID returns the wire protocol packet identifier.
func (p CurrencyBalancePacket) PacketID() uint16 { return CurrencyResponsePacketID }

// Encode serializes the activity-point balances as a length-prefixed sequence of type/amount pairs.
func (p CurrencyBalancePacket) Encode() ([]byte, error) {
	w := codec.NewWriter()
	w.WriteInt32(int32(len(p.Currencies)))
	for _, c := range p.Currencies {
		w.WriteInt32(int32(c.Type))
		w.WriteInt32(int32(c.Amount))
	}
	return w.Bytes(), nil
}
