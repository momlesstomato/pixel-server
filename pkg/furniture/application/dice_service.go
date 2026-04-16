package application

import (
	"context"
	"strconv"

	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// StartDiceRoll marks one placed dice item as rolling.
func (service *Service) StartDiceRoll(ctx context.Context, itemID, roomID int) (domain.Item, domain.Definition, bool, error) {
	item, def, err := service.loadPlacedItemDefinition(ctx, itemID, roomID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, false, err
	}
	if def.InteractionType != domain.InteractionDice {
		return domain.Item{}, domain.Definition{}, false, domain.ErrInvalidInteraction
	}
	if item.ExtraData == "-1" {
		return item, def, false, nil
	}
	item.ExtraData = "-1"
	if err := service.repository.UpdateItemData(ctx, item.ID, item.ExtraData); err != nil {
		return domain.Item{}, domain.Definition{}, false, err
	}
	return item, def, true, nil
}

// FinishDiceRoll persists one final dice result.
func (service *Service) FinishDiceRoll(ctx context.Context, itemID, roomID, value int) (domain.Item, domain.Definition, error) {
	item, def, err := service.loadPlacedItemDefinition(ctx, itemID, roomID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	if def.InteractionType != domain.InteractionDice || value < 1 || value > 6 {
		return domain.Item{}, domain.Definition{}, domain.ErrInvalidInteraction
	}
	item.ExtraData = strconv.Itoa(value)
	if err := service.repository.UpdateItemData(ctx, item.ID, item.ExtraData); err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	return item, def, nil
}

// ClearDice clears one placed dice item back to its idle state.
func (service *Service) ClearDice(ctx context.Context, itemID, roomID int) (domain.Item, domain.Definition, error) {
	item, def, err := service.loadPlacedItemDefinition(ctx, itemID, roomID)
	if err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	if def.InteractionType != domain.InteractionDice {
		return domain.Item{}, domain.Definition{}, domain.ErrInvalidInteraction
	}
	item.ExtraData = "0"
	if err := service.repository.UpdateItemData(ctx, item.ID, item.ExtraData); err != nil {
		return domain.Item{}, domain.Definition{}, err
	}
	return item, def, nil
}
