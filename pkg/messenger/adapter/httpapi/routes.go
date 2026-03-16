package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
)

// RegisterRoutes registers all messenger admin REST API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("messenger service is required")
	}
	registerFriendRoutes(module, service)
	registerRequestRoutes(module, service)
	registerRelationshipRoutes(module, service)
	return nil
}

// registerFriendRoutes registers friend add/remove endpoints.
func registerFriendRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/users/:id/friends", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		friends, findErr := service.ListFriends(ctx.UserContext(), userID)
		if findErr != nil {
			return fiber.NewError(http.StatusInternalServerError, findErr.Error())
		}
		return ctx.JSON(mapFriendships(friends))
	})
	module.RegisterPOST("/api/v1/users/:id/friends", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload addFriendRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		if addErr := service.AddFriendship(ctx.UserContext(), userID, payload.FriendID); addErr != nil {
			return mapMessengerError(addErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
	module.RegisterDELETE("/api/v1/users/:id/friends/:friendId", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		friendID, err := parsePositiveID(ctx.Params("friendId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if delErr := service.RemoveFriendship(ctx.UserContext(), userID, friendID); delErr != nil {
			return mapMessengerError(delErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
}
