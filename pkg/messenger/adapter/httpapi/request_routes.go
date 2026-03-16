package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/messenger/domain"
)

// registerRequestRoutes registers the pending-request listing endpoint.
func registerRequestRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id/friends/requests", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		requests, listErr := service.ListPendingRequests(ctx.UserContext(), userID)
		if listErr != nil {
			return fiber.NewError(http.StatusInternalServerError, listErr.Error())
		}
		return ctx.JSON(mapRequests(requests))
	})
}

// registerRelationshipRoutes registers get and patch relationship endpoints.
func registerRelationshipRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id/friends/:friendId/relationship", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		_, err = parsePositiveID(ctx.Params("friendId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		counts, cErr := service.GetRelationshipCounts(ctx.UserContext(), userID)
		if cErr != nil {
			return fiber.NewError(http.StatusInternalServerError, cErr.Error())
		}
		return ctx.JSON(fiber.Map{"counts": counts})
	})
	module.RegisterPATCH("/api/v1/users/:id/friends/:friendId/relationship", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		friendID, err := parsePositiveID(ctx.Params("friendId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload relationshipPatchRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		rel := domain.RelationshipType(payload.Type)
		if setErr := service.SetRelationship(ctx.UserContext(), userID, friendID, rel); setErr != nil {
			return mapMessengerError(setErr)
		}
		return ctx.JSON(fiber.Map{"type": payload.Type})
	})
}
