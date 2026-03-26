package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/economy/domain"
)

// RegisterRoutes registers economy API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("economy service is required")
	}
	registerMarketplaceRoutes(module, service)
	return nil
}

// registerMarketplaceRoutes registers marketplace offer routes.
func registerMarketplaceRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/marketplace/offers", func(ctx *fiber.Ctx) error {
		filter := domain.OfferFilter{
			MinPrice: ctx.QueryInt("min_price"),
			MaxPrice: ctx.QueryInt("max_price"),
			Offset:   ctx.QueryInt("offset"),
			Limit:    ctx.QueryInt("limit", 20),
		}
		offers, total, err := service.ListOpenOffers(ctx.UserContext(), filter)
		if err != nil {
			return mapEconomyError(err)
		}
		return ctx.JSON(fiber.Map{"offers": offers, "total": total})
	})
	module.RegisterGET("/api/v1/marketplace/offers/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		offer, findErr := service.FindOfferByID(ctx.UserContext(), id)
		if findErr != nil {
			return mapEconomyError(findErr)
		}
		return ctx.JSON(offer)
	})
	module.RegisterPOST("/api/v1/marketplace/offers", func(ctx *fiber.Ctx) error {
		var payload domain.MarketplaceOffer
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		created, createErr := service.CreateOffer(ctx.UserContext(), payload)
		if createErr != nil {
			return mapEconomyError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(created)
	})
	module.RegisterDELETE("/api/v1/marketplace/offers/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if cancelErr := service.CancelOffer(ctx.UserContext(), id); cancelErr != nil {
			return mapEconomyError(cancelErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
	module.RegisterGET("/api/v1/marketplace/history/:spriteId", func(ctx *fiber.Ctx) error {
		spriteID, err := parsePositiveID(ctx.Params("spriteId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		history, findErr := service.GetPriceHistory(ctx.UserContext(), spriteID)
		if findErr != nil {
			return mapEconomyError(findErr)
		}
		return ctx.JSON(history)
	})
	module.RegisterGET("/api/v1/marketplace/sellers/:sellerId/offers", func(ctx *fiber.Ctx) error {
		sellerID, err := parsePositiveID(ctx.Params("sellerId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		offers, findErr := service.ListOffersBySellerID(ctx.UserContext(), sellerID)
		if findErr != nil {
			return mapEconomyError(findErr)
		}
		return ctx.JSON(offers)
	})
}

// parsePositiveID validates and parses a positive integer identifier.
func parsePositiveID(value string) (int, error) {
	id, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("id must be a positive integer")
	}
	return id, nil
}

// mapEconomyError maps domain errors to HTTP responses.
func mapEconomyError(err error) error {
	switch err {
	case domain.ErrOfferNotFound, domain.ErrTradeLogNotFound:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrOfferNotOpen, domain.ErrMarketplaceDisabled:
		return fiber.NewError(http.StatusConflict, err.Error())
	case domain.ErrSelfPurchase, domain.ErrMaxOffersReached, domain.ErrItemNotMarketable:
		return fiber.NewError(http.StatusForbidden, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
