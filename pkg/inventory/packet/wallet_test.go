package packet_test

import (
	"encoding/binary"
	"testing"

	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
	"github.com/momlesstomato/pixel-server/pkg/inventory/packet"
)

// TestCreditBalancePacketID verifies the packet uses the credits response identifier.
func TestCreditBalancePacketID(t *testing.T) {
	pkt := packet.CreditBalancePacket{Balance: 100}
	if pkt.PacketID() != packet.CreditsResponsePacketID {
		t.Fatalf("expected CreditsResponsePacketID, got %d", pkt.PacketID())
	}
}

// TestCreditBalancePacketEncode verifies credits are encoded as the string form.
func TestCreditBalancePacketEncode(t *testing.T) {
	body, err := packet.CreditBalancePacket{Balance: 250}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	length := binary.BigEndian.Uint16(body[:2])
	str := string(body[2 : 2+length])
	if str != "250.0" {
		t.Fatalf("expected 250.0, got %q", str)
	}
}

// TestCurrencyBalancePacketID verifies the packet uses the currency response identifier.
func TestCurrencyBalancePacketID(t *testing.T) {
	pkt := packet.CurrencyBalancePacket{}
	if pkt.PacketID() != packet.CurrencyResponsePacketID {
		t.Fatalf("expected CurrencyResponsePacketID, got %d", pkt.PacketID())
	}
}

// TestCurrencyBalancePacketEncodeEmpty verifies an empty list encodes as count zero.
func TestCurrencyBalancePacketEncodeEmpty(t *testing.T) {
	body, err := packet.CurrencyBalancePacket{}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	count := int32(binary.BigEndian.Uint32(body[:4]))
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}

// TestCurrencyBalancePacketEncodeEntries verifies type/amount pairs are serialized correctly.
func TestCurrencyBalancePacketEncodeEntries(t *testing.T) {
	currencies := []domain.Currency{
		{Type: domain.CurrencyDuckets, Amount: 100},
		{Type: domain.CurrencyDiamonds, Amount: 5},
	}
	body, err := packet.CurrencyBalancePacket{Currencies: currencies}.Encode()
	if err != nil {
		t.Fatalf("unexpected encode error: %v", err)
	}
	count := int(int32(binary.BigEndian.Uint32(body[0:4])))
	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}
	type0 := int32(binary.BigEndian.Uint32(body[4:8]))
	amt0 := int32(binary.BigEndian.Uint32(body[8:12]))
	if type0 != int32(domain.CurrencyDuckets) || amt0 != 100 {
		t.Fatalf("unexpected first entry type=%d amount=%d", type0, amt0)
	}
	type1 := int32(binary.BigEndian.Uint32(body[12:16]))
	amt1 := int32(binary.BigEndian.Uint32(body[16:20]))
	if type1 != int32(domain.CurrencyDiamonds) || amt1 != 5 {
		t.Fatalf("unexpected second entry type=%d amount=%d", type1, amt1)
	}
}
