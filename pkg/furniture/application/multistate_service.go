package application

import (
	"context"
	"strconv"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// ToggleMultistate advances one placed multistate item to its next configured mode.
func (service *Service) ToggleMultistate(ctx context.Context, itemID, roomID int) (domain.Item, domain.Definition, error) {
	item, def, err := service.loadPlacedItemDefinition(ctx, itemID, roomID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	if def.InteractionModesCount < 2 {
		return domain.Item{}, domain.Definition{}, domain.ErrInvalidInteraction
	}
	next := (parseModeState(item.ExtraData, def.InteractionModesCount) + 1) % def.InteractionModesCount
	item.ExtraData = strconv.Itoa(next)
	if err := service.repository.UpdateItemData(ctx, item.ID, item.ExtraData); err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	return item, def, nil
}

// loadPlacedItemDefinition resolves one placed item and its definition in a target room.
func (service *Service) loadPlacedItemDefinition(ctx context.Context, itemID, roomID int) (domain.Item, domain.Definition, error) {
	item, err := service.FindItemByID(ctx, itemID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	if item.RoomID == 0 {
		return domain.Item{}, domain.Definition{}, domain.ErrItemNotPlaced
	}
	if roomID > 0 && item.RoomID != roomID {
		return domain.Item{}, domain.Definition{}, domain.ErrItemNotFound
	}
	def, err := service.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	return item, def, nil
}

// parseModeState resolves one bounded multistate value and falls back to zero on malformed data.
func parseModeState(raw string, modes int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 || value >= modes {
		return 0
	}
	return value
}
