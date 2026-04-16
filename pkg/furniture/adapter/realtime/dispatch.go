package realtime

import (
	"context"
	"strconv"
	"strings"

	"github.com/momlesstomato/pixel-server/core/codec"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	"go.uber.org/zap"
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
	case furnipacket.PostItPlacePacketID:
		return true, runtime.handlePostItPlace(ctx, connID, body)
	case furnipacket.PickupPacketID:
		return true, runtime.handlePickup(ctx, connID, body)
	case furnipacket.FloorUpdatePacketID:
		return true, runtime.handleFloorUpdate(ctx, connID, body)
	case furnipacket.WallUpdatePacketID:
		return true, runtime.handleWallUpdate(ctx, connID, body)
	case furnipacket.ToggleMultistatePacketID:
		return true, runtime.handleToggleMultistate(ctx, connID, body)
	case furnipacket.ToggleWallMultistatePacketID:
		return true, runtime.handleToggleWallMultistate(ctx, connID, body)
	case furnipacket.ActivateDicePacketID:
		return true, runtime.handleActivateDice(ctx, connID, body)
	case furnipacket.DeactivateDicePacketID:
		return true, runtime.handleDeactivateDice(ctx, connID, body)
	case furnipacket.SetStackHeightPacketID:
		return true, runtime.handleSetStackHeight(ctx, connID, body)
	case furnipacket.GetItemDataPacketID:
		return true, runtime.handleGetItemData(ctx, connID, body)
	case furnipacket.SetItemDataPacketID:
		return true, runtime.handleSetItemData(ctx, connID, body)
	case furnipacket.DimmerSettingsPacketID:
		return true, runtime.handleDimmerSettings(ctx, connID)
	case furnipacket.DimmerSavePacketID:
		return true, runtime.handleDimmerSave(ctx, connID, body)
	case furnipacket.DimmerTogglePacketID:
		return true, runtime.handleDimmerToggle(ctx, connID)
	case furnipacket.OpenPresentPacketID:
		return true, runtime.handleOpenPresent(ctx, connID, body)
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
	if runtime.roomFinder == nil || runtime.roomBroadcaster == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	r := codec.NewReader(body)
	raw, err := r.ReadString()
	if err != nil {
		return nil
	}
	raw = strings.TrimSpace(raw)
	splitAt := strings.IndexByte(raw, ' ')
	if splitAt <= 0 {
		return nil
	}
	itemID, _ := strconv.Atoi(raw[:splitAt])
	if itemID <= 0 {
		return nil
	}
	if !runtime.canManageItem(ctx, roomID, userID, itemID) {
		return nil
	}
	existing, err := runtime.service.FindItemByID(ctx, itemID)
	if err != nil {
		runtime.logger.Warn("find place item failed", zap.Int("item_id", itemID), zap.Error(err))
		return nil
	}
	if existing.RoomID != 0 {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, existing.DefinitionID)
	if err != nil {
		runtime.logger.Warn("find place item definition failed", zap.Int("definition_id", existing.DefinitionID), zap.Error(err))
		return nil
	}
	placement := strings.TrimSpace(raw[splitAt+1:])
	if def.ItemType == furnituredomain.ItemTypeWall {
		return runtime.placeWallItem(ctx, connID, roomID, userID, existing, def, placement)
	}
	parts := strings.Fields(placement)
	if len(parts) < 3 {
		return nil
	}
	x, _ := strconv.Atoi(parts[0])
	y, _ := strconv.Atoi(parts[1])
	dir, _ := strconv.Atoi(parts[2])
	if runtime.isTileOccupied(roomID, x, y) {
		return nil
	}
	item, err := runtime.service.PlaceFloorItem(ctx, itemID, userID, roomID, x, y, dir)
	if err != nil {
		runtime.logger.Warn("place floor item failed", zap.Int("item_id", itemID), zap.Int("room_id", roomID), zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	pkt := furnipacket.FloorItemAddPacket{
		ItemID: item.ID, SpriteID: def.SpriteID,
		X: item.X, Y: item.Y, Z: item.Z, Dir: item.Dir,
		StackHeight: runtime.effectiveStackHeight(item, def),
		ExtraData:   item.ExtraData, UserID: item.UserID,
	}
	encoded, err := pkt.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemAddPacketID, encoded)
	runtime.syncFloorItemEntries(roomID, item, def)
	return runtime.sendUserPacket(ctx, connID, item.UserID, furnipacket.InventoryRemovePacket{ItemID: item.ID})
}

// handlePostItPlace processes a sticky-note wall placement request (c2s 2248).
func (runtime *Runtime) handlePostItPlace(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	if runtime.roomFinder == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	wallPosition, err := r.ReadString()
	if err != nil {
		return nil
	}
	existing, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, existing.DefinitionID)
	if err != nil {
		return nil
	}
	return runtime.placeWallItem(ctx, connID, roomID, userID, existing, def, wallPosition)
}

// placeWallItem places one wall item and broadcasts the room delta.
func (runtime *Runtime) placeWallItem(ctx context.Context, connID string, roomID int, userID int, existing furnituredomain.Item, def furnituredomain.Definition, wallPosition string) error {
	if !runtime.canManageItem(ctx, roomID, userID, existing.ID) || strings.TrimSpace(wallPosition) == "" {
		return nil
	}
	item, err := runtime.service.PlaceWallItem(ctx, existing.ID, userID, roomID, wallPosition)
	if err != nil {
		runtime.logger.Warn("place wall item failed", zap.Int("item_id", existing.ID), zap.Int("room_id", roomID), zap.Int("user_id", userID), zap.Error(err))
		return nil
	}
	username := ""
	if runtime.usernameResolver != nil {
		if resolved, resolveErr := runtime.usernameResolver(ctx, item.UserID); resolveErr == nil {
			username = resolved
		}
	}
	body, err := furnipacket.WallItemAddPacket{Item: runtime.wallItemPacket(item, def), Username: username}.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.WallItemAddPacketID, body)
	return runtime.sendUserPacket(ctx, connID, item.UserID, furnipacket.InventoryRemovePacket{ItemID: item.ID})
}
