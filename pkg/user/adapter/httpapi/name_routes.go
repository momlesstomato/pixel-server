package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/user/domain"
)

// registerNameRoutes registers user rename routes.
func registerNameRoutes(module *corehttp.Module, service Service) {
	module.RegisterPOST("/api/v1/users/:id/name-change", func(ctx *fiber.Ctx) error {
		userID, err := parseUserID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload nameChangeRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		result, changeErr := service.ForceChangeName(ctx.UserContext(), userID, payload.Name)
		if changeErr != nil {
			return mapUserError(changeErr)
		}
		if result.ResultCode != domain.NameResultAvailable {
			return fiber.NewError(http.StatusConflict, "name change rejected")
		}
		return ctx.JSON(result)
	})
}

// nameChangeRequest defines administrative rename request payload.
type nameChangeRequest struct {
	// Name stores requested username value.
	Name string `json:"name"`
}
