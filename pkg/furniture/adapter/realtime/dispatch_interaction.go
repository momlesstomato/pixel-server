package realtime

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/momlesstomato/pixel-server/core/codec"
	furnitureapp "github.com/momlesstomato/pixel-server/pkg/furniture/application"
	furnituredomain "github.com/momlesstomato/pixel-server/pkg/furniture/domain"
	furnipacket "github.com/momlesstomato/pixel-server/pkg/furniture/packet"
	"go.uber.org/zap"
)

const (
	teleporterTransferDelay = 650 * time.Millisecond
	teleporterExitDelay     = 500 * time.Millisecond
	teleporterResetDelay    = 1000 * time.Millisecond
)

// handleToggleMultistate processes a furniture state toggle (c2s 99).
func (runtime *Runtime) handleToggleMultistate(ctx context.Context, connID string, body []byte) error {
	itemID, roomID, ok := runtime.itemInteractionContext(connID, body)
	if !ok {
		return nil
	}
	item, err := runtime.service.FindItemByID(ctx, itemID)
	if err != nil || item.RoomID != roomID {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return nil
	}
	if def.InteractionType == furnituredomain.InteractionTeleport {
		return runtime.handleUseTeleporter(ctx, connID, roomID, item, def)
	}
	if !runtime.canUseFloorItem(roomID, item, connID) {
		return nil
	}
	updatedItem, updatedDef, err := runtime.service.ToggleMultistate(ctx, itemID, roomID)
	if err != nil {
		return nil
	}
	return runtime.broadcastFloorItemState(roomID, updatedItem, updatedDef)
}

// handleToggleWallMultistate processes a wall furniture state toggle (c2s 210).
func (runtime *Runtime) handleToggleWallMultistate(ctx context.Context, connID string, body []byte) error {
	if runtime.roomFinder == nil {
		return nil
	}
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	item, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil || item.RoomID != roomID {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || def.ItemType != furnituredomain.ItemTypeWall {
		return nil
	}
	updated, updatedDef, err := runtime.service.ToggleMultistate(ctx, int(itemID), roomID)
	if err != nil {
		return nil
	}
	return runtime.broadcastWallItemState(roomID, updated, updatedDef)
}

// handleActivateDice processes a dice roll request (c2s 1990).
func (runtime *Runtime) handleActivateDice(ctx context.Context, connID string, body []byte) error {
	itemID, roomID, ok := runtime.itemInteractionContext(connID, body)
	if !ok {
		return nil
	}
	if _, ok := runtime.resolveUsableFloorItem(ctx, itemID, roomID, connID); !ok {
		return nil
	}
	item, def, started, err := runtime.service.StartDiceRoll(ctx, itemID, roomID)
	if err != nil {
		return nil
	}
	if err := runtime.broadcastFloorItemState(roomID, item, def); err != nil {
		return err
	}
	if started {
		runtime.scheduleDiceResolution(roomID, itemID)
	}
	return runtime.broadcastDiceValue(roomID, itemID, -1)
}

// handleDeactivateDice processes a dice reset request (c2s 1533).
func (runtime *Runtime) handleDeactivateDice(ctx context.Context, connID string, body []byte) error {
	itemID, roomID, ok := runtime.itemInteractionContext(connID, body)
	if !ok {
		return nil
	}
	if _, ok := runtime.resolveUsableFloorItem(ctx, itemID, roomID, connID); !ok {
		return nil
	}
	item, def, err := runtime.service.ClearDice(ctx, itemID, roomID)
	if err != nil {
		return nil
	}
	runtime.cancelDiceResolution(itemID)
	if err := runtime.broadcastFloorItemState(roomID, item, def); err != nil {
		return err
	}
	return runtime.broadcastDiceValue(roomID, itemID, 0)
}

// handleSetStackHeight processes a stack-helper height change (c2s 3839).
func (runtime *Runtime) handleSetStackHeight(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil || runtime.roomFinder == nil {
		return nil
	}
	heightRaw, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok || !runtime.canManageItem(ctx, roomID, userID, int(itemID)) {
		return nil
	}
	item, def, err := runtime.service.SetStackHeight(ctx, int(itemID), roomID, float64(heightRaw)/100)
	if err != nil {
		return nil
	}
	if err := runtime.broadcastFloorItemState(roomID, item, def); err != nil {
		return err
	}
	return runtime.sendPacket(connID, furnipacket.StackHeightUpdatePacket{ItemID: itemID, Height: int32(math.Round(runtime.effectiveStackHeight(item, def) * 100))})
}

// handleGetItemData processes a wall item data request (c2s 3964).
func (runtime *Runtime) handleGetItemData(ctx context.Context, connID string, body []byte) error {
	if runtime.roomFinder == nil {
		return nil
	}
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return nil
	}
	item, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil || item.RoomID != roomID {
		return nil
	}
	return runtime.sendPacket(connID, furnipacket.ItemDataUpdatePacket{ItemID: item.ID, Data: item.InteractionData})
}

// handleSetItemData processes a sticky-note save request (c2s 3666).
func (runtime *Runtime) handleSetItemData(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	if runtime.roomFinder == nil {
		return nil
	}
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	color, err := r.ReadString()
	if err != nil {
		return nil
	}
	text, err := r.ReadString()
	if err != nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok || !runtime.canModifyRoom(ctx, roomID, userID) {
		return nil
	}
	item, err := runtime.service.FindItemByID(ctx, int(itemID))
	if err != nil || item.RoomID != roomID {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || def.InteractionType != furnituredomain.InteractionPostIt {
		return nil
	}
	if len(text) > 500 {
		text = text[:500]
	}
	color = runtime.normalizePostItColor(color)
	if _, err := runtime.service.UpdateItemInteractionData(ctx, item.ID, text); err != nil {
		return nil
	}
	updated, err := runtime.service.UpdateItemData(ctx, item.ID, color)
	if err != nil {
		return nil
	}
	updated.InteractionData = text
	if err := runtime.broadcastWallItemState(roomID, updated, def); err != nil {
		return err
	}
	return runtime.broadcastWallItemData(roomID, updated.ID, text)
}

// handleDimmerSettings processes a room dimmer preset request (c2s 2813).
func (runtime *Runtime) handleDimmerSettings(ctx context.Context, connID string) error {
	userID, _ := runtime.userID(connID)
	if runtime.roomFinder == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok || !runtime.canModifyRoom(ctx, roomID, userID) {
		return nil
	}
	_, _, data, ok := runtime.findRoomDimmer(ctx, roomID)
	if !ok {
		return nil
	}
	presets := make([]furnipacket.DimmerPreset, 0, len(data.Presets))
	for _, preset := range data.Presets {
		presets = append(presets, furnipacket.DimmerPreset{PresetID: preset.PresetID, Type: preset.Type, Color: preset.Color, Brightness: preset.Brightness})
	}
	return runtime.sendPacket(connID, furnipacket.DimmerPresetsPacket{SelectedPresetID: data.SelectedPresetID, Presets: presets})
}

// handleDimmerSave processes a room dimmer save request (c2s 1648).
func (runtime *Runtime) handleDimmerSave(ctx context.Context, connID string, body []byte) error {
	userID, _ := runtime.userID(connID)
	if runtime.roomFinder == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok || !runtime.canModifyRoom(ctx, roomID, userID) {
		return nil
	}
	r := codec.NewReader(body)
	presetID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	effectID, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	color, err := r.ReadString()
	if err != nil {
		return nil
	}
	brightness, err := r.ReadInt32()
	if err != nil {
		return nil
	}
	apply, err := r.ReadBool()
	if err != nil {
		return nil
	}
	item, def, data, ok := runtime.findRoomDimmer(ctx, roomID)
	if !ok {
		return nil
	}
	data.SelectedPresetID = int(presetID)
	updated := false
	for index := range data.Presets {
		if data.Presets[index].PresetID != int(presetID) {
			continue
		}
		data.Presets[index].Type = int(effectID)
		data.Presets[index].Color = runtime.normalizeDimmerColor(color)
		data.Presets[index].Brightness = runtime.normalizeDimmerBrightness(int(brightness))
		updated = true
		break
	}
	if !updated {
		data.Presets = append(data.Presets, furnituredomain.DimmerPresetData{PresetID: int(presetID), Type: int(effectID), Color: runtime.normalizeDimmerColor(color), Brightness: runtime.normalizeDimmerBrightness(int(brightness))})
	}
	if apply {
		data.Enabled = true
	}
	item, err = runtime.persistDimmerState(ctx, item, data)
	if err != nil {
		return nil
	}
	if err := runtime.broadcastWallItemState(roomID, item, def); err != nil {
		return err
	}
	return runtime.handleDimmerSettings(ctx, connID)
}

// handleDimmerToggle processes a room dimmer toggle request (c2s 2296).
func (runtime *Runtime) handleDimmerToggle(ctx context.Context, connID string) error {
	userID, _ := runtime.userID(connID)
	if runtime.roomFinder == nil {
		return nil
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok || !runtime.canModifyRoom(ctx, roomID, userID) {
		return nil
	}
	item, def, data, ok := runtime.findRoomDimmer(ctx, roomID)
	if !ok {
		return nil
	}
	data.Enabled = !data.Enabled
	item, err := runtime.persistDimmerState(ctx, item, data)
	if err != nil {
		return nil
	}
	if err := runtime.broadcastWallItemState(roomID, item, def); err != nil {
		return err
	}
	return runtime.handleDimmerSettings(ctx, connID)
}

// handleOpenPresent processes a present opening request (c2s 3558).
func (runtime *Runtime) handleOpenPresent(ctx context.Context, connID string, body []byte) error {
	itemID, roomID, ok := runtime.itemInteractionContext(connID, body)
	if !ok {
		return nil
	}
	item, ok := runtime.resolveUsableFloorItem(ctx, itemID, roomID, connID)
	if !ok {
		return nil
	}
	def, err := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || def.InteractionType != furnituredomain.InteractionGift {
		return nil
	}
	metadata, err := furnituredomain.ParseInteractionData(item.InteractionData)
	if err != nil || metadata.Gift == nil || metadata.Gift.DefinitionID <= 0 {
		return nil
	}
	revealedDef, err := runtime.service.FindDefinitionByID(ctx, metadata.Gift.DefinitionID)
	if err != nil {
		return nil
	}
	transformed, err := runtime.service.TransformItem(ctx, item.ID, revealedDef.ID, "0", "")
	if err != nil {
		return nil
	}
	if revealedDef.ItemType == furnituredomain.ItemTypeFloor {
		if err := runtime.broadcastFloorItemState(roomID, transformed, revealedDef); err != nil {
			return err
		}
	}
	productCode := metadata.Gift.ProductCode
	if productCode == "" {
		productCode = revealedDef.ItemName
	}
	return runtime.sendPacket(connID, furnipacket.GiftOpenedPacket{
		ItemType: string(revealedDef.ItemType), ClassID: revealedDef.SpriteID,
		ProductCode: productCode, PlacedItemID: transformed.ID,
		PlacedItemType: string(revealedDef.ItemType), PlacedInRoom: transformed.RoomID != 0,
		PetFigureString: metadata.Gift.PetFigureString,
	})
}

// itemInteractionContext resolves one item interaction payload and active room pairing.
func (runtime *Runtime) itemInteractionContext(connID string, body []byte) (int, int, bool) {
	r := codec.NewReader(body)
	itemID, err := r.ReadInt32()
	if err != nil || runtime.roomFinder == nil {
		return 0, 0, false
	}
	roomID, ok := runtime.roomFinder(connID)
	if !ok {
		return 0, 0, false
	}
	return int(itemID), roomID, true
}

// canUseFloorItem reports whether the actor is close enough to use one placed floor item.
func (runtime *Runtime) canUseFloorItem(roomID int, item furnituredomain.Item, connID string) bool {
	if runtime.roomEntityTileResolver == nil {
		return true
	}
	activeRoomID, x, y, ok := runtime.roomEntityTileResolver(connID)
	if !ok || activeRoomID != roomID {
		return false
	}
	return abs(item.X-x) <= 1 && abs(item.Y-y) <= 1
}

// resolveUsableFloorItem validates room membership and proximity for one floor item interaction.
func (runtime *Runtime) resolveUsableFloorItem(ctx context.Context, itemID, roomID int, connID string) (furnituredomain.Item, bool) {
	item, err := runtime.service.FindItemByID(ctx, itemID)
	if err != nil || item.RoomID != roomID || !runtime.canUseFloorItem(roomID, item, connID) {
		return furnituredomain.Item{}, false
	}
	return item, true
}

// broadcastFloorItemState emits one floor item update to all occupants in the room.
func (runtime *Runtime) broadcastFloorItemState(roomID int, item furnituredomain.Item, def furnituredomain.Definition) error {
	if runtime.roomBroadcaster == nil {
		return nil
	}
	body, err := furnipacket.FloorItemUpdatePacket{ItemID: item.ID, SpriteID: def.SpriteID, X: item.X, Y: item.Y, Z: item.Z, Dir: item.Dir, StackHeight: runtime.effectiveStackHeight(item, def), ExtraData: item.ExtraData, UserID: item.UserID}.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.FloorItemUpdatePacketID, body)
	return nil
}

// broadcastWallItemState emits one wall item update to all occupants in the room.
func (runtime *Runtime) broadcastWallItemState(roomID int, item furnituredomain.Item, def furnituredomain.Definition) error {
	if runtime.roomBroadcaster == nil {
		return nil
	}
	body, err := furnipacket.WallItemUpdatePacket{Item: runtime.wallItemPacket(item, def)}.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.WallItemUpdatePacketID, body)
	return nil
}

// broadcastWallItemData emits one hidden wall item data update.
func (runtime *Runtime) broadcastWallItemData(roomID int, itemID int, data string) error {
	if runtime.roomBroadcaster == nil {
		return nil
	}
	body, err := furnipacket.ItemDataUpdatePacket{ItemID: itemID, Data: data}.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.ItemDataUpdatePacketID, body)
	return nil
}

// broadcastDiceValue emits one dice animation or final value update.
func (runtime *Runtime) broadcastDiceValue(roomID, itemID, value int) error {
	if runtime.roomBroadcaster == nil {
		return nil
	}
	body, err := furnipacket.DiceValuePacket{ItemID: int32(itemID), Value: int32(value)}.Encode()
	if err != nil {
		return err
	}
	runtime.roomBroadcaster(roomID, furnipacket.DiceValuePacketID, body)
	return nil
}

// effectiveStackHeight resolves the client-visible height for one placed item.
func (runtime *Runtime) effectiveStackHeight(item furnituredomain.Item, def furnituredomain.Definition) float64 {
	return furnitureapp.EffectiveStackHeight(item, def)
}

// scheduleDiceResolution finalises one rolling dice after the configured animation delay.
func (runtime *Runtime) scheduleDiceResolution(roomID, itemID int) {
	runtime.cancelDiceResolution(itemID)
	ctx, cancel := context.WithCancel(context.Background())
	runtime.diceMu.Lock()
	runtime.diceCancels[itemID] = cancel
	runtime.diceMu.Unlock()
	go func() {
		timer := time.NewTimer(runtime.diceRollDelay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		runtime.cancelDiceResolution(itemID)
		value := 1
		if runtime.diceRandomizer != nil {
			value = runtime.diceRandomizer(6) + 1
		}
		if value < 1 || value > 6 {
			value = 1
		}
		item, def, err := runtime.service.FinishDiceRoll(context.Background(), itemID, roomID, value)
		if err != nil {
			runtime.logger.Warn("finish dice roll failed", zap.Int("item_id", itemID), zap.Int("room_id", roomID), zap.Error(err))
			return
		}
		if err := runtime.broadcastFloorItemState(roomID, item, def); err != nil {
			runtime.logger.Warn("broadcast dice item state failed", zap.Int("item_id", itemID), zap.Error(err))
			return
		}
		if err := runtime.broadcastDiceValue(roomID, itemID, value); err != nil {
			runtime.logger.Warn("broadcast dice value failed", zap.Int("item_id", itemID), zap.Error(err))
		}
	}()
}

// cancelDiceResolution stops one in-flight dice roll finaliser.
func (runtime *Runtime) cancelDiceResolution(itemID int) {
	runtime.diceMu.Lock()
	cancel, ok := runtime.diceCancels[itemID]
	if ok {
		delete(runtime.diceCancels, itemID)
	}
	runtime.diceMu.Unlock()
	if ok {
		cancel()
	}
}

// abs resolves the absolute integer distance helper used by proximity checks.
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

// wallItemPacket converts one wall item pair into its wire payload.
func (runtime *Runtime) wallItemPacket(item furnituredomain.Item, def furnituredomain.Definition) furnipacket.FurnitureWallItem {
	return furnipacket.FurnitureWallItem{ItemID: item.ID, SpriteID: def.SpriteID, WallPosition: item.WallPosition, ExtraData: item.ExtraData, UserID: item.UserID}
}

// normalizePostItColor clamps sticky-note colors to the supported palette.
func (runtime *Runtime) normalizePostItColor(value string) string {
	upper := strings.ToUpper(strings.TrimSpace(strings.TrimPrefix(value, "#")))
	switch upper {
	case "9CCEFF", "FF9CFF", "9CFF9C", "FFFF33":
		return upper
	default:
		return "FFFF33"
	}
}

// normalizeDimmerColor clamps room dimmer colors to a hashed RGB payload.
func (runtime *Runtime) normalizeDimmerColor(value string) string {
	color := strings.ToUpper(strings.TrimSpace(value))
	if !strings.HasPrefix(color, "#") {
		color = "#" + strings.TrimPrefix(color, "#")
	}
	if len(color) != 7 {
		return "#000000"
	}
	return color
}

// normalizeDimmerBrightness clamps room dimmer brightness values.
func (runtime *Runtime) normalizeDimmerBrightness(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

// findRoomDimmer resolves the first placed room dimmer and its decoded metadata.
func (runtime *Runtime) findRoomDimmer(ctx context.Context, roomID int) (furnituredomain.Item, furnituredomain.Definition, furnituredomain.DimmerData, bool) {
	items, err := runtime.service.ListRoomItems(ctx, roomID)
	if err != nil {
		return furnituredomain.Item{}, furnituredomain.Definition{}, furnituredomain.DimmerData{}, false
	}
	for _, item := range items {
		def, defErr := runtime.service.FindDefinitionByID(ctx, item.DefinitionID)
		if defErr != nil || def.InteractionType != furnituredomain.InteractionDimmer {
			continue
		}
		data := runtime.loadDimmerData(item.InteractionData)
		return item, def, data, true
	}
	return furnituredomain.Item{}, furnituredomain.Definition{}, furnituredomain.DimmerData{}, false
}

// loadDimmerData decodes one room dimmer metadata payload with defaults.
func (runtime *Runtime) loadDimmerData(raw string) furnituredomain.DimmerData {
	data := furnituredomain.DimmerData{Enabled: false, SelectedPresetID: 1, Presets: []furnituredomain.DimmerPresetData{{PresetID: 1, Type: 1, Color: "#000000", Brightness: 255}, {PresetID: 2, Type: 1, Color: "#000000", Brightness: 255}, {PresetID: 3, Type: 1, Color: "#000000", Brightness: 255}}}
	parsed, err := furnituredomain.ParseInteractionData(raw)
	if err != nil || parsed.Dimmer == nil {
		return data
	}
	if parsed.Dimmer.SelectedPresetID > 0 {
		data.SelectedPresetID = parsed.Dimmer.SelectedPresetID
	}
	data.Enabled = parsed.Dimmer.Enabled
	if len(parsed.Dimmer.Presets) > 0 {
		data.Presets = parsed.Dimmer.Presets
	}
	return data
}

// persistDimmerState stores room dimmer metadata and visible state.
func (runtime *Runtime) persistDimmerState(ctx context.Context, item furnituredomain.Item, data furnituredomain.DimmerData) (furnituredomain.Item, error) {
	raw, err := furnituredomain.InteractionData{Dimmer: &data}.Encode()
	if err != nil {
		return furnituredomain.Item{}, err
	}
	if _, err := runtime.service.UpdateItemInteractionData(ctx, item.ID, raw); err != nil {
		return furnituredomain.Item{}, err
	}
	selected := data.Presets[0]
	for _, preset := range data.Presets {
		if preset.PresetID == data.SelectedPresetID {
			selected = preset
			break
		}
	}
	enabled := 0
	if data.Enabled {
		enabled = 1
	}
	updated, err := runtime.service.UpdateItemData(ctx, item.ID, strings.Join([]string{strconv.Itoa(enabled), strconv.Itoa(data.SelectedPresetID), strconv.Itoa(selected.Type), selected.Color, strconv.Itoa(selected.Brightness)}, ","))
	if err != nil {
		return furnituredomain.Item{}, err
	}
	updated.InteractionData = raw
	return updated, nil
}

// handleUseTeleporter starts the teleporter transfer sequence for one item.
func (runtime *Runtime) handleUseTeleporter(ctx context.Context, connID string, roomID int, item furnituredomain.Item, def furnituredomain.Definition) error {
	if item.ExtraData != "" && item.ExtraData != "0" {
		return nil
	}
	actor, ok := runtime.findRoomEntityByConn(roomID, connID)
	if !ok {
		return nil
	}
	useX, useY := runtime.frontTile(item.X, item.Y, item.Dir)
	if (actor.X != useX || actor.Y != useY) && (actor.X != item.X || actor.Y != item.Y) {
		if runtime.roomEntityWalker == nil {
			return nil
		}
		if err := runtime.roomEntityWalker(ctx, connID, useX, useY); err != nil {
			return nil
		}
		go func(virtualID int) {
			if runtime.waitForEntityTile(connID, roomID, useX, useY, 3*time.Second) {
				_ = runtime.executeTeleporterTransfer(connID, roomID, virtualID, item, def)
			}
		}(actor.VirtualID)
		return nil
	}
	go func(virtualID int) {
		_ = runtime.executeTeleporterTransfer(connID, roomID, virtualID, item, def)
	}(actor.VirtualID)
	return nil
}

// executeTeleporterTransfer opens the source and destination teleporters and moves the actor.
func (runtime *Runtime) executeTeleporterTransfer(connID string, roomID int, sourceVirtualID int, item furnituredomain.Item, def furnituredomain.Definition) error {
	partner, partnerDef, ok := runtime.resolveTeleporterPartner(context.Background(), item)
	if !ok || partner.RoomID == 0 {
		return nil
	}
	updatedSource, err := runtime.updateFloorItemState(context.Background(), roomID, item, def, "1")
	if err == nil {
		item = updatedSource
	}
	if sourceVirtualID > 0 && runtime.roomEntityWarper != nil {
		_ = runtime.roomEntityWarper(context.Background(), roomID, sourceVirtualID, item.X, item.Y, item.Z, item.Dir, false, true)
	}
	exitX, exitY := runtime.frontTile(partner.X, partner.Y, partner.Dir)
	if partner.RoomID != roomID {
		go func(sourceItem furnituredomain.Item, targetItem furnituredomain.Item, targetDef furnituredomain.Definition, targetExitX int, targetExitY int) {
			transferTimer := time.NewTimer(teleporterTransferDelay)
			defer transferTimer.Stop()
			<-transferTimer.C
			if updatedSourceItem, updateErr := runtime.updateFloorItemState(context.Background(), roomID, sourceItem, def, "2"); updateErr == nil {
				sourceItem = updatedSourceItem
			}
			if updatedPartner, updateErr := runtime.updateFloorItemState(context.Background(), targetItem.RoomID, targetItem, targetDef, "2"); updateErr == nil {
				targetItem = updatedPartner
			}
			if runtime.teleporterForwarder != nil {
				_ = runtime.teleporterForwarder(context.Background(), connID, targetItem.RoomID, targetItem.X, targetItem.Y, targetItem.Z, targetItem.Dir, targetExitX, targetExitY)
			}
			if runtime.waitForEntityTile(connID, targetItem.RoomID, targetItem.X, targetItem.Y, 3*time.Second) {
				exitTimer := time.NewTimer(teleporterExitDelay)
				defer exitTimer.Stop()
				<-exitTimer.C
				if updatedPartner, updateErr := runtime.updateFloorItemState(context.Background(), targetItem.RoomID, targetItem, targetDef, "1"); updateErr == nil {
					targetItem = updatedPartner
				}
				if runtime.roomEntityWalker != nil && (targetExitX != targetItem.X || targetExitY != targetItem.Y) {
					_ = runtime.roomEntityWalker(context.Background(), connID, targetExitX, targetExitY)
				}
			}
			runtime.scheduleTeleporterReset(sourceItem.ID, roomID, targetItem.ID, targetItem.RoomID, teleporterResetDelay)
		}(item, partner, partnerDef, exitX, exitY)
		return nil
	}
	if sourceVirtualID == 0 {
		runtime.scheduleTeleporterReset(item.ID, roomID, partner.ID, partner.RoomID, teleporterTransferDelay+teleporterResetDelay)
		return nil
	}
	go func(sourceItem furnituredomain.Item, targetItem furnituredomain.Item, targetDef furnituredomain.Definition, virtualID int, targetExitX int, targetExitY int) {
		transferTimer := time.NewTimer(teleporterTransferDelay)
		defer transferTimer.Stop()
		<-transferTimer.C
		if updatedSourceItem, updateErr := runtime.updateFloorItemState(context.Background(), roomID, sourceItem, def, "2"); updateErr == nil {
			sourceItem = updatedSourceItem
		}
		if updatedPartner, updateErr := runtime.updateFloorItemState(context.Background(), targetItem.RoomID, targetItem, targetDef, "2"); updateErr == nil {
			targetItem = updatedPartner
		}
		if runtime.roomEntityWarper != nil {
			_ = runtime.roomEntityWarper(context.Background(), roomID, virtualID, targetItem.X, targetItem.Y, targetItem.Z, targetItem.Dir, false, false)
		}
		exitTimer := time.NewTimer(teleporterExitDelay)
		defer exitTimer.Stop()
		<-exitTimer.C
		if updatedPartner, updateErr := runtime.updateFloorItemState(context.Background(), targetItem.RoomID, targetItem, targetDef, "1"); updateErr == nil {
			targetItem = updatedPartner
		}
		if runtime.roomEntityWalker != nil && (targetExitX != targetItem.X || targetExitY != targetItem.Y) {
			_ = runtime.roomEntityWalker(context.Background(), connID, targetExitX, targetExitY)
		}
		runtime.scheduleTeleporterReset(sourceItem.ID, roomID, targetItem.ID, targetItem.RoomID, teleporterResetDelay)
	}(item, partner, partnerDef, sourceVirtualID, exitX, exitY)
	return nil
}

// resolveTeleporterPartner resolves or auto-pairs one teleporter destination.
func (runtime *Runtime) resolveTeleporterPartner(ctx context.Context, item furnituredomain.Item) (furnituredomain.Item, furnituredomain.Definition, bool) {
	metadata, err := furnituredomain.ParseInteractionData(item.InteractionData)
	if err == nil && metadata.Teleporter != nil && metadata.Teleporter.ItemID > 0 {
		partner, findErr := runtime.service.FindItemByID(ctx, metadata.Teleporter.ItemID)
		if findErr == nil {
			item, partner = runtime.syncTeleporterMetadata(ctx, item, partner)
			def, defErr := runtime.service.FindDefinitionByID(ctx, partner.DefinitionID)
			if defErr == nil {
				return partner, def, true
			}
		}
	}
	items, err := runtime.service.ListItemsByUserID(ctx, item.UserID)
	if err != nil {
		return furnituredomain.Item{}, furnituredomain.Definition{}, false
	}
	for _, candidate := range items {
		if candidate.ID == item.ID || candidate.DefinitionID != item.DefinitionID {
			continue
		}
		candidateMeta, parseErr := furnituredomain.ParseInteractionData(candidate.InteractionData)
		if parseErr == nil && candidateMeta.Teleporter != nil && candidateMeta.Teleporter.ItemID != item.ID {
			continue
		}
		_, updatedCandidate := runtime.syncTeleporterMetadata(ctx, item, candidate)
		candidateDef, defErr := runtime.service.FindDefinitionByID(ctx, updatedCandidate.DefinitionID)
		if defErr != nil {
			return furnituredomain.Item{}, furnituredomain.Definition{}, false
		}
		return updatedCandidate, candidateDef, true
	}
	return furnituredomain.Item{}, furnituredomain.Definition{}, false
}

// scheduleTeleporterReset closes the teleporter state after a short delay.
func (runtime *Runtime) scheduleTeleporterReset(sourceID int, sourceRoomID int, targetID int, targetRoomID int, delay time.Duration) {
	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		<-timer.C
		source, err := runtime.service.FindItemByID(context.Background(), sourceID)
		if err == nil {
			def, defErr := runtime.service.FindDefinitionByID(context.Background(), source.DefinitionID)
			if defErr == nil {
				if updated, updateErr := runtime.updateFloorItemState(context.Background(), sourceRoomID, source, def, "0"); updateErr == nil {
					source = updated
				}
			}
		}
		target, err := runtime.service.FindItemByID(context.Background(), targetID)
		if err == nil {
			def, defErr := runtime.service.FindDefinitionByID(context.Background(), target.DefinitionID)
			if defErr == nil {
				if updated, updateErr := runtime.updateFloorItemState(context.Background(), targetRoomID, target, def, "0"); updateErr == nil {
					target = updated
				}
			}
		}
	}()
}

// updateFloorItemState persists one floor item state string and broadcasts the update.
func (runtime *Runtime) updateFloorItemState(ctx context.Context, roomID int, item furnituredomain.Item, def furnituredomain.Definition, state string) (furnituredomain.Item, error) {
	updated, err := runtime.service.UpdateItemData(ctx, item.ID, state)
	if err != nil {
		return furnituredomain.Item{}, err
	}
	if err := runtime.broadcastFloorItemState(roomID, updated, def); err != nil {
		return updated, err
	}
	return updated, nil
}

// syncTeleporterMetadata keeps persisted teleporter room and item references aligned after moves.
func (runtime *Runtime) syncTeleporterMetadata(ctx context.Context, left furnituredomain.Item, right furnituredomain.Item) (furnituredomain.Item, furnituredomain.Item) {
	leftData, leftErr := furnituredomain.ParseInteractionData(left.InteractionData)
	rightData, rightErr := furnituredomain.ParseInteractionData(right.InteractionData)
	leftValid := leftErr == nil && leftData.Teleporter != nil && leftData.Teleporter.ItemID == right.ID && leftData.Teleporter.RoomID == right.RoomID
	rightValid := rightErr == nil && rightData.Teleporter != nil && rightData.Teleporter.ItemID == left.ID && rightData.Teleporter.RoomID == left.RoomID
	if leftValid && rightValid {
		return left, right
	}
	leftRaw, leftEncodeErr := furnituredomain.InteractionData{Teleporter: &furnituredomain.TeleporterData{RoomID: right.RoomID, ItemID: right.ID}}.Encode()
	rightRaw, rightEncodeErr := furnituredomain.InteractionData{Teleporter: &furnituredomain.TeleporterData{RoomID: left.RoomID, ItemID: left.ID}}.Encode()
	if leftEncodeErr != nil || rightEncodeErr != nil {
		return left, right
	}
	if updatedLeft, updateErr := runtime.service.UpdateItemInteractionData(ctx, left.ID, leftRaw); updateErr == nil {
		left = updatedLeft
	}
	if updatedRight, updateErr := runtime.service.UpdateItemInteractionData(ctx, right.ID, rightRaw); updateErr == nil {
		right = updatedRight
	}
	return left, right
}

// findRoomEntityByConn resolves one room entity snapshot by connection identifier.
func (runtime *Runtime) findRoomEntityByConn(roomID int, connID string) (RoomEntitySnapshot, bool) {
	if runtime.roomEntitySnapshotter == nil {
		return RoomEntitySnapshot{}, false
	}
	for _, entity := range runtime.roomEntitySnapshotter(roomID) {
		if entity.ConnID == connID {
			return entity, true
		}
	}
	return RoomEntitySnapshot{}, false
}

// waitForEntityTile waits until one actor reaches the target tile or times out.
func (runtime *Runtime) waitForEntityTile(connID string, roomID int, x int, y int, timeout time.Duration) bool {
	if runtime.roomEntityTileResolver == nil {
		return false
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timer.C:
			return false
		case <-ticker.C:
			activeRoomID, tileX, tileY, ok := runtime.roomEntityTileResolver(connID)
			if ok && activeRoomID == roomID && tileX == x && tileY == y {
				return true
			}
		}
	}
}

// frontTile resolves one tile immediately in front of an item direction.
func (runtime *Runtime) frontTile(x int, y int, dir int) (int, int) {
	switch dir {
	case 1, 2, 3:
		return x + 1, y
	case 5, 6, 7:
		return x - 1, y
	case 4:
		return x, y + 1
	default:
		return x, y - 1
	}
}

// rollerIsActive reports whether one roller should advance during the room tick.
func (runtime *Runtime) rollerIsActive(item furnituredomain.Item, def furnituredomain.Definition) bool {
	if def.InteractionModesCount < 2 {
		return true
	}
	return runtime.parseMultistateValue(item.ExtraData, def.InteractionModesCount) > 0
}

// parseMultistateValue decodes one bounded multistate integer from item extra data.
func (runtime *Runtime) parseMultistateValue(raw string, modes int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value >= modes {
		return 0
	}
	return value
}

// rollerTargetBaseHeight resolves the static target tile height used for a roller move.
func (runtime *Runtime) rollerTargetBaseHeight(roomID, x, y int, fallback float64) (float64, bool) {
	if runtime.roomTileResolver == nil {
		return fallback, true
	}
	tile, ok := runtime.roomTileResolver(roomID, x, y)
	if !ok || !tile.Walkable {
		return 0, false
	}
	return tile.Z, true
}

// rollerTargetOccupied reports whether one unmoved avatar already occupies a destination tile.
func (runtime *Runtime) rollerTargetOccupied(x, y int, entities []RoomEntitySnapshot, movedUnits map[int]bool) bool {
	for _, entity := range entities {
		if movedUnits[entity.VirtualID] {
			continue
		}
		if entity.X == x && entity.Y == y {
			return true
		}
	}
	return false
}

// ProcessRoomTick advances autonomous roller interactions for one room.
func (runtime *Runtime) ProcessRoomTick(roomID int) {
	if runtime.roomBroadcaster == nil || runtime.roomEntitySnapshotter == nil {
		return
	}
	items, err := runtime.service.ListRoomItems(context.Background(), roomID)
	if err != nil {
		return
	}
	definitions := make(map[int]furnituredomain.Definition)
	tileItems := make(map[[2]int][]furnituredomain.Item)
	rollers := make([]furnituredomain.Item, 0)
	for _, item := range items {
		def, defErr := runtime.service.FindDefinitionByID(context.Background(), item.DefinitionID)
		if defErr != nil || def.ItemType != furnituredomain.ItemTypeFloor {
			continue
		}
		definitions[item.DefinitionID] = def
		key := [2]int{item.X, item.Y}
		tileItems[key] = append(tileItems[key], item)
		if def.InteractionType == furnituredomain.InteractionRoller && runtime.rollerIsActive(item, def) {
			rollers = append(rollers, item)
		}
	}
	if len(rollers) == 0 {
		return
	}
	movedItems := make(map[int]bool)
	movedUnits := make(map[int]bool)
	entities := runtime.roomEntitySnapshotter(roomID)
	for _, roller := range rollers {
		runtime.processRoller(roomID, roller, definitions, tileItems, entities, movedItems, movedUnits)
	}
}

// processRoller advances one roller tile worth of movements.
func (runtime *Runtime) processRoller(roomID int, roller furnituredomain.Item, definitions map[int]furnituredomain.Definition, tileItems map[[2]int][]furnituredomain.Item, entities []RoomEntitySnapshot, movedItems map[int]bool, movedUnits map[int]bool) {
	rollerDef, ok := definitions[roller.DefinitionID]
	if !ok || rollerDef.InteractionType != furnituredomain.InteractionRoller {
		return
	}
	targetX, targetY := runtime.frontTile(roller.X, roller.Y, roller.Dir)
	if targetX == roller.X && targetY == roller.Y {
		return
	}
	if runtime.TileBlockCheckerFor(roomID, targetX, targetY) {
		return
	}
	if runtime.rollerTargetOccupied(targetX, targetY, entities, movedUnits) {
		return
	}
	itemKey := [2]int{roller.X, roller.Y}
	targetKey := [2]int{targetX, targetY}
	var movingItem furnituredomain.Item
	var movingDef furnituredomain.Definition
	itemFound := false
	for _, candidate := range tileItems[itemKey] {
		def, exists := definitions[candidate.DefinitionID]
		if !exists || candidate.ID == roller.ID || movedItems[candidate.ID] || def.InteractionType == furnituredomain.InteractionRoller {
			continue
		}
		if !itemFound || candidate.Z >= movingItem.Z {
			movingItem = candidate
			movingDef = def
			itemFound = true
		}
	}
	var movingUnit RoomEntitySnapshot
	unitFound := false
	for _, entity := range entities {
		if movedUnits[entity.VirtualID] || entity.X != roller.X || entity.Y != roller.Y || entity.IsWalking {
			continue
		}
		movingUnit = entity
		unitFound = true
		break
	}
	if !itemFound && !unitFound {
		return
	}
	nextHeight, ok := runtime.rollerTargetBaseHeight(roomID, targetX, targetY, roller.Z)
	if !ok {
		return
	}
	for _, candidate := range tileItems[targetKey] {
		def, exists := definitions[candidate.DefinitionID]
		if !exists || movedItems[candidate.ID] || candidate.ID == roller.ID {
			continue
		}
		height := candidate.Z + runtime.effectiveStackHeight(candidate, def)
		if height > nextHeight {
			nextHeight = height
		}
	}
	packet := furnipacket.RoomRollingPacket{SourceX: roller.X, SourceY: roller.Y, TargetX: targetX, TargetY: targetY, RollerID: roller.ID}
	if itemFound {
		updated, err := runtime.service.MovePlacedItem(context.Background(), movingItem.ID, roomID, targetX, targetY, nextHeight, movingItem.Dir)
		if err == nil {
			packet.Items = append(packet.Items, furnipacket.RollingItem{ItemID: movingItem.ID, Height: movingItem.Z, NextHeight: updated.Z})
			movedItems[movingItem.ID] = true
			if movingDef.CanSit || movingDef.CanLay {
				runtime.replaceSeatEntries(roomID, updated.ID, seatEntriesFromFootprint(updated.ID, updated.X, updated.Y, updated.Dir, runtime.effectiveStackHeight(updated, movingDef), movingDef.Width, movingDef.Length, movingDef.CanSit, movingDef.CanLay))
			}
			if shouldBlockFloorItem(movingDef) {
				runtime.replaceBlockEntries(roomID, updated.ID, blockEntriesFromFootprint(updated.ID, updated.X, updated.Y, updated.Dir, movingDef.Width, movingDef.Length))
			} else {
				runtime.removeBlockEntries(roomID, updated.ID)
			}
		}
	}
	if unitFound && runtime.roomEntityWarper != nil {
		if err := runtime.roomEntityWarper(context.Background(), roomID, movingUnit.VirtualID, targetX, targetY, nextHeight, movingUnit.Dir, true, false); err == nil {
			packet.Unit = &furnipacket.RollingUnit{MovementType: 2, UnitID: movingUnit.VirtualID, Height: movingUnit.Z, NextHeight: nextHeight}
			movedUnits[movingUnit.VirtualID] = true
		}
	}
	if len(packet.Items) == 0 && packet.Unit == nil {
		return
	}
	body, err := packet.Encode()
	if err != nil {
		return
	}
	runtime.roomBroadcaster(roomID, furnipacket.RoomRollingPacketID, body)
	updatedTargetItems := make([]furnituredomain.Item, 0, len(tileItems[targetKey])+1)
	for _, candidate := range tileItems[targetKey] {
		updatedTargetItems = append(updatedTargetItems, candidate)
	}
	if itemFound {
		movingItem.X = targetX
		movingItem.Y = targetY
		movingItem.Z = nextHeight
		updatedTargetItems = append(updatedTargetItems, movingItem)
	}
	tileItems[targetKey] = updatedTargetItems
}
