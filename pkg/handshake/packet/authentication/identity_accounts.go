package authentication

import (
	"fmt"

	"github.com/momlesstomato/pixel-server/core/codec"
)

// IdentityAccountsPacketID identifies handshake.identity_accounts packet.
const IdentityAccountsPacketID uint16 = 3523

// IdentityAccount defines one selectable identity account entry.
type IdentityAccount struct {
	// ID stores account identifier.
	ID int32
	// Name stores account display name.
	Name string
}

// IdentityAccountsPacket carries identity account rows.
type IdentityAccountsPacket struct {
	// Accounts stores all account entries available to the client.
	Accounts []IdentityAccount
}

// PacketID returns protocol packet id.
func (packet IdentityAccountsPacket) PacketID() uint16 { return IdentityAccountsPacketID }

// Decode parses packet body into struct fields.
func (packet *IdentityAccountsPacket) Decode(body []byte) error {
	reader := codec.NewReader(body)
	count, err := reader.ReadInt32()
	if err != nil {
		return err
	}
	if count < 0 {
		return fmt.Errorf("identity account count must be non-negative")
	}
	accounts := make([]IdentityAccount, 0, count)
	for index := int32(0); index < count; index++ {
		accountID, accountIDErr := reader.ReadInt32()
		if accountIDErr != nil {
			return accountIDErr
		}
		accountName, accountNameErr := reader.ReadString()
		if accountNameErr != nil {
			return accountNameErr
		}
		accounts = append(accounts, IdentityAccount{ID: accountID, Name: accountName})
	}
	if reader.Remaining() != 0 {
		return fmt.Errorf("identity_accounts body has %d trailing bytes", reader.Remaining())
	}
	packet.Accounts = accounts
	return nil
}

// Encode serializes packet fields into protocol body bytes.
func (packet IdentityAccountsPacket) Encode() ([]byte, error) {
	writer := codec.NewWriter()
	writer.WriteInt32(int32(len(packet.Accounts)))
	for _, account := range packet.Accounts {
		writer.WriteInt32(account.ID)
		if err := writer.WriteString(account.Name); err != nil {
			return nil, err
		}
	}
	return writer.Bytes(), nil
}
