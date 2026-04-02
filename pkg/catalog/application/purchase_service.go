package application

import (
	"context"
	"fmt"

	sdkcatalog "github.com/momlesstomato/pixel-sdk/events/catalog"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// PurchaseOffer validates and charges a catalog offer purchase for one user.
// Credits and activity points are deducted via the configured Spender port.
// Limited-edition sell counts are incremented atomically.
// Returns the purchased offer so the caller may notify the client.
func (service *Service) PurchaseOffer(ctx context.Context, connID string, userID int, offerID int, extraData string, amount int) (domain.CatalogOffer, error) {
	if amount < 1 {
		return domain.CatalogOffer{}, fmt.Errorf("purchase amount must be positive")
	}
	offer, err := service.repository.FindOfferByID(ctx, offerID)
	if err != nil {
		return domain.CatalogOffer{}, err
	}
	if !offer.OfferActive {
		return domain.CatalogOffer{}, domain.ErrOfferInactive
	}
	if err := service.chargePurchase(ctx, userID, offer, amount); err != nil {
		return domain.CatalogOffer{}, err
	}
	if service.fire != nil {
		ev := &sdkcatalog.OfferPurchased{ConnID: connID, UserID: userID, OfferID: offer.ID, Quantity: amount}
		service.fire(ev)
		if ev.Cancelled() {
			return domain.CatalogOffer{}, fmt.Errorf("purchase cancelled by plugin")
		}
	}
	if offer.IsLimited() {
		ok, incErr := service.repository.IncrementLimitedSells(ctx, offer.ID)
		if incErr != nil {
			return domain.CatalogOffer{}, incErr
		}
		if !ok {
			return domain.CatalogOffer{}, domain.ErrOfferSoldOut
		}
	}
	if service.fire != nil {
		service.fire(&sdkcatalog.OfferPurchaseConfirmed{ConnID: connID, UserID: userID, OfferID: offer.ID, Quantity: amount})
	}
	return offer, nil
}

// PurchaseGift validates and charges a catalog gift purchase sent to a recipient.
// The offer is purchased from the actor's balance; delivery is fire-and-forget.
func (service *Service) PurchaseGift(ctx context.Context, connID string, actorUserID int, offerID int, extraData string, recipientUsername string) (domain.CatalogOffer, error) {
	if service.recipientFinder == nil {
		return domain.CatalogOffer{}, domain.ErrRecipientNotFound
	}
	recipient, err := service.recipientFinder.FindRecipientByUsername(ctx, recipientUsername)
	if err != nil {
		return domain.CatalogOffer{}, domain.ErrRecipientNotFound
	}
	if !recipient.AllowGifts {
		return domain.CatalogOffer{}, domain.ErrRecipientNotFound
	}
	return service.PurchaseOffer(ctx, connID, actorUserID, offerID, extraData, 1)
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
