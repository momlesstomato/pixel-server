package realtime

import (
	"context"

	userdomain "github.com/momlesstomato/pixel-server/pkg/user/domain"
	packetname "github.com/momlesstomato/pixel-server/pkg/user/packet/name"
)

// handleCheckName processes one user.check_name packet.
func (runtime *Runtime) handleCheckName(ctx context.Context, connID string, userID int, body []byte) error {
	packet := packetname.UserNameInputPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packetname.UserCheckNamePacketID, err)
	}
	result, err := runtime.service.CheckName(ctx, packet.Name, userID)
	if err != nil {
		return runtime.logError(connID, packetname.UserCheckNamePacketID, err)
	}
	response := packetname.UserNameResultPacket{ResultCode: result.ResultCode, Name: result.Name, Suggestions: result.Suggestions}
	return runtime.transportResult(connID, packetname.UserCheckNameResultPacketID, response)
}

// handleChangeName processes one user.change_name and user.approve_name packet.
func (runtime *Runtime) handleChangeName(ctx context.Context, connID string, userID int, body []byte, force bool) error {
	packet := packetname.UserNameInputPacket{}
	if err := packet.Decode(body); err != nil {
		return runtime.logError(connID, packetname.UserChangeNamePacketID, err)
	}
	result, err := runtime.service.ChangeName(ctx, connID, userID, packet.Name, force)
	if err != nil {
		return runtime.logError(connID, packetname.UserChangeNamePacketID, err)
	}
	response := packetname.UserNameResultPacket{ResultCode: result.ResultCode, Name: result.Name, Suggestions: result.Suggestions}
	if err := runtime.transportResult(connID, packetname.UserChangeNameResultPacketID, response); err != nil {
		return err
	}
	if result.ResultCode != userdomain.NameResultAvailable {
		return nil
	}
	change := packetname.UserNameChangePacket{WebID: int32(userID), UserID: int32(userID), NewName: result.Name}
	return runtime.sendPacket(connID, change)
}

// transportResult sends a name-result packet using an explicit packet identifier.
func (runtime *Runtime) transportResult(connID string, packetID uint16, packet packetname.UserNameResultPacket) error {
	body, err := packet.Encode()
	if err != nil {
		return err
	}
	return runtime.transport.Send(connID, packetID, body)
}
