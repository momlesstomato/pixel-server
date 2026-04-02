package httpapi

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/inventory/domain"
)

// registerCurrencyRoutes registers currency balance routes.
func registerCurrencyRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/inventory/:userId/credits", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		credits, findErr := service.GetCredits(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(fiber.Map{"credits": credits})
	})
	module.RegisterPOST("/api/v1/inventory/:userId/credits", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload currencyModifyRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		newBalance, addErr := service.AddCredits(ctx.UserContext(), userID, payload.Amount)
		if addErr != nil {
			return mapInventoryError(addErr)
		}
		return ctx.JSON(fiber.Map{"credits": newBalance})
	})
	module.RegisterGET("/api/v1/inventory/:userId/currencies", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		currencies, findErr := service.ListCurrencies(ctx.UserContext(), userID)
		if findErr != nil {
			return mapInventoryError(findErr)
		}
		return ctx.JSON(currencies)
	})
	module.RegisterPOST("/api/v1/inventory/:userId/currencies/:type", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		ct, err := parseCurrencyTypeID(ctx.Params("type"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload currencyModifyRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		src := domain.TransactionSource(payload.Source)
		if src == "" {
			src = domain.SourceAdmin
		}
		newBalance, addErr := service.AddCurrencyTracked(ctx.UserContext(), userID, domain.CurrencyType(ct), payload.Amount, src, "api", "")
		if addErr != nil {
			return mapInventoryError(addErr)
		}
		return ctx.JSON(fiber.Map{"balance": newBalance})
	})
}

// currencyModifyRequest defines currency modification payload.
type currencyModifyRequest struct {
	// Amount stores the signed currency delta.
	Amount int `json:"amount"`
	// Source stores the transaction source identifier.
	Source string `json:"source"`
}
