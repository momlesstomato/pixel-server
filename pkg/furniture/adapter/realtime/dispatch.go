package realtime

import (
	"context"
	"strconv"
	"strings"

	"github.com/momlesstomato/pixel-server/core/codec"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
)

// Handle dispatches one authenticated furniture packet payload.
func (runtime *Runtime) Handle(ctx context.Context, connID string, packetID uint16, body []byte) (bool, error) {
	_, ok := runtime.userID(connID)
	if !ok {
		return false, nil
	}
	switch packetID {
	case furnipacket.GetFurniturePacketID:
		return true, runtime.handleGetFurniture(ctx, connID)
	case furnipacket.PlacePacketID:
		return true, runtime.handlePlace(ctx, connID, body)
	case furnipacket.PickupPacketID:
		return true, runtime.handlePickup(ctx, connID, body)
	case furnipacket.FloorUpdatePacketID:
		return true, runtime.handleFloorUpdate(ctx, connID, body)
	case furnipacket.ToggleMultistatePacketID:
		return true, runtime.handleToggleMultistate(ctx, connID, body)
	default:
		return false, nil
	}
}

// handleGetFurniture responds with the user's full furniture inventory.
func (runtime *Runtime) handleGetFurniture(ctx context.Context, connID string) error {
	userID, _ := runtime.userID(connID)
	items, err := runtime.service.ListItemsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	entries := make([]furnipacket.FurniListItem, 0, len(items))
	for _, item := range items {
		if item.RoomID != 0 {
			continue
		}
		def, defErr := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
		if defErr != nil {
			continue
		}
		entries = append(entries, furnipacket.FurniListItem{
			ID:                   item.ID,
			ItemType:             def.ItemType,
			SpriteID:             def.SpriteID,
			ExtraData:            item.ExtraData,
			LimitedNumber:        item.LimitedNumber,
			LimitedTotal:         item.LimitedTotal,
			AllowRecycle:         def.AllowRecycle,
			AllowTrade:           def.AllowTrade,
			AllowInventoryStack:  def.AllowInventoryStack,
			AllowMarketplaceSell: def.AllowMarketplaceSell,
		})
	}
	return runtime.sendPacket(connID, furnipacket.FurniListPacket{
		TotalFragments: 1,
		FragmentIndex:  0,
		Items:          entries,
	})
}

// handlePlace processes a furniture placement request.
func (runtime *Runtime) handlePlace(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	r := codec.NewReader(body)
	raw, err := r.ReadString()
	if err != nil {
		return nil
	}
	parts := strings.Fields(raw)
	if len(parts) < 4 {
		return nil
	}
	itemID, _ := strconv.Atoi(parts[0])
	x, _ := strconv.Atoi(parts[1])
	y, _ := strconv.Atoi(parts[2])
	dir, _ := strconv.Atoi(parts[3])
	if itemID <= 0 {
		return nil
	}
	if runtime.roomFinder == nil || runtime.roomBroadcaster == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	item, err := runtime.service.PlaceFloorItem(ctx, itemID, userID, roomID, x, y, dir)
	if err != nil {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return nil
	}
	pkt := furnipacket.FloorItemAddPacket{
		ItemID: item.ID, SpriteID: def.SpriteID,
		X: item.X, Y: item.Y, Z: item.Z, Dir: item.Dir,
		ExtraData: item.ExtraData, UserID: item.UserID,
	}
	encoded, err := pkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemAddPacketID, encoded)
	if def.CanSit {
		runtime.addSeatEntry(roomID, item.ID, item.X, item.Y, item.Dir, def.StackHeight, true)
	}
	return runtime.sendPacket(connID, furnipacket.InventoryRemovePacket{ItemID: item.ID})
}
