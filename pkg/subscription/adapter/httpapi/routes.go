package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/subscription/domain"
)

// RegisterRoutes registers subscription API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("subscription service is required")
	}
	registerSubscriptionRoutes(module, service)
	registerClubOfferRoutes(module, service)
	return nil
}

// registerSubscriptionRoutes registers subscription query routes.
func registerSubscriptionRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/subscriptions/user/:userId", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		sub, findErr := service.FindActiveSubscription(ctx.UserContext(), userID)
		if findErr != nil {
			return mapSubscriptionError(findErr)
		}
		return ctx.JSON(sub)
	})
}

// registerClubOfferRoutes registers club offer CRUD routes.
func registerClubOfferRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/subscriptions/offers", func(ctx *fiber.Ctx) error {
		offers, err := service.ListClubOffers(ctx.UserContext())
		if err != nil {
			return mapSubscriptionError(err)
		}
		return ctx.JSON(offers)
	})
	module.RegisterGET("/api/v1/subscriptions/offers/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		offer, findErr := service.FindClubOfferByID(ctx.UserContext(), id)
		if findErr != nil {
			return mapSubscriptionError(findErr)
		}
		return ctx.JSON(offer)
	})
	module.RegisterPOST("/api/v1/subscriptions/offers", func(ctx *fiber.Ctx) error {
		var payload domain.ClubOffer
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		created, createErr := service.CreateClubOffer(ctx.UserContext(), payload)
		if createErr != nil {
			return mapSubscriptionError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(created)
	})
	module.RegisterDELETE("/api/v1/subscriptions/offers/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if deleteErr := service.DeleteClubOffer(ctx.UserContext(), id); deleteErr != nil {
			return mapSubscriptionError(deleteErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
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

// mapSubscriptionError maps domain errors to HTTP responses.
func mapSubscriptionError(err error) error {
	switch err {
	case domain.ErrSubscriptionNotFound, domain.ErrClubOfferNotFound:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrPurchaseLimitReached:
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
