package realtime

import (
	"context"

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
	_ = body
	return nil
}

// handlePickup processes a furniture pickup request.
func (runtime *Runtime) handlePickup(ctx context.Context, connID string, body []byte) error {
	_ = body
	return nil
}

// handleToggleMultistate processes a furniture state toggle.
func (runtime *Runtime) handleToggleMultistate(ctx context.Context, connID string, body []byte) error {
	_ = body
	return nil
}
