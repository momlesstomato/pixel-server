package application

import (
	"context"
	"fmt"

	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// PurchaseOffer validates, charges and delivers a catalog offer purchase for one user.
// Credits and activity points are deducted via the configured Spender port.
// A furniture item is created via the optional ItemDeliverer port when the offer has a
// linked item definition.
// Limited-edition sell counts are incremented atomically after charge.
func (service *Service) PurchaseOffer(ctx context.Context, connID string, userID int, offerID int, extraData string, amount int) (domain.PurchaseResult, error) {
	if amount < 1 {
		return domain.PurchaseResult{}, fmt.Errorf("purchase amount must be positive")
	}
	offer, err := service.repository.FindOfferByID(ctx, offerID)
	if err != nil {
		return domain.PurchaseResult{}, err
	}
	if !offer.OfferActive {
		return domain.PurchaseResult{}, domain.ErrOfferInactive
	}
	if err := service.chargePurchase(ctx, userID, offer, amount); err != nil {
		return domain.PurchaseResult{}, err
	}
	var newCredits int
	if service.spender != nil {
		var creditsErr error
		newCredits, creditsErr = service.spender.GetCredits(ctx, userID)
		if creditsErr != nil {
			return domain.PurchaseResult{}, creditsErr
		}
	}
	if service.fire != nil {
		ev := &sdkcatalog.OfferPurchased{ConnID: connID, UserID: userID, OfferID: offer.ID, Quantity: amount}
		service.fire(ev)
		if ev.Cancelled() {
			return domain.PurchaseResult{}, fmt.Errorf("purchase cancelled by plugin")
		}
	}
	if offer.IsLimited() {
		ok, incErr := service.repository.IncrementLimitedSells(ctx, offer.ID)
		if incErr != nil {
			return domain.PurchaseResult{}, incErr
		}
		if !ok {
			return domain.PurchaseResult{}, domain.ErrOfferSoldOut
		}
	}
	var itemID int
	if service.itemDeliverer != nil && offer.ItemDefinitionID > 0 {
		var deliverErr error
		itemID, deliverErr = service.itemDeliverer.DeliverItem(ctx, userID, offer.ItemDefinitionID, extraData, 0, 0)
		if deliverErr != nil {
			return domain.PurchaseResult{}, deliverErr
		}
	}
	if service.fire != nil {
		service.fire(&sdkcatalog.OfferPurchaseConfirmed{ConnID: connID, UserID: userID, OfferID: offer.ID, Quantity: amount})
	}
	return domain.PurchaseResult{Offer: offer, ItemID: itemID, NewCredits: newCredits}, nil
}

// PurchaseGift validates and charges a catalog gift purchase sent to a recipient.
// The offer is charged from the actor's balance; the item is delivered to the recipient.
func (service *Service) PurchaseGift(ctx context.Context, connID string, actorUserID int, offerID int, extraData string, recipientUsername string) (domain.PurchaseResult, error) {
	if service.recipientFinder == nil {
		return domain.PurchaseResult{}, domain.ErrRecipientNotFound
	}
	recipient, err := service.recipientFinder.FindRecipientByUsername(ctx, recipientUsername)
	if err != nil {
		return domain.PurchaseResult{}, domain.ErrRecipientNotFound
	}
	if !recipient.AllowGifts {
		return domain.PurchaseResult{}, domain.ErrRecipientNotFound
	}
	result, purchaseErr := service.PurchaseOffer(ctx, connID, actorUserID, offerID, extraData, 1)
	if purchaseErr != nil {
		return domain.PurchaseResult{}, purchaseErr
	}
	if service.itemDeliverer != nil && result.Offer.ItemDefinitionID > 0 && recipient.UserID != actorUserID {
		var giftDeliverErr error
		result.ItemID, giftDeliverErr = service.itemDeliverer.DeliverItem(ctx, recipient.UserID, result.Offer.ItemDefinitionID, extraData, 0, 0)
		if giftDeliverErr != nil {
			return domain.PurchaseResult{}, giftDeliverErr
		}
	}
	return result, nil
}

// chargePurchase deducts credits and activity points for one offer purchase.
func (service *Service) chargePurchase(ctx context.Context, userID int, offer domain.CatalogOffer, amount int) error {
	totalCredits := offer.CostCredits * amount
	totalPoints := offer.CostActivityPoints * amount
	if totalCredits <= 0 && totalPoints <= 0 {
		return nil
	}
	if service.spender == nil {
		return domain.ErrNoSpender
	}
	if totalCredits > 0 {
		balance, err := service.spender.GetCredits(ctx, userID)
		if err != nil {
			return err
		}
		if balance < totalCredits {
			return domain.ErrInsufficientCredits
		}
		if _, err := service.spender.AddCredits(ctx, userID, -totalCredits); err != nil {
			return err
		}
	}
	if totalPoints > 0 {
		balance, err := service.spender.GetCurrencyBalance(ctx, userID, offer.ActivityPointType)
		if err != nil {
			return err
		}
		if balance < totalPoints {
			return domain.ErrInsufficientActivityPoints
		}
		if _, err := service.spender.AddCurrencyBalance(ctx, userID, offer.ActivityPointType, -totalPoints); err != nil {
			return err
		}
	}
	return nil
}
