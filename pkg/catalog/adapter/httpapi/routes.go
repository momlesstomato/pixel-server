package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/catalog/domain"
)

// RegisterRoutes registers catalog API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("catalog service is required")
	}
	registerPageRoutes(module, service)
	registerOfferRoutes(module, service)
	registerVoucherRoutes(module, service)
	return nil
}

// registerPageRoutes registers catalog page CRUD routes.
func registerPageRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/catalog/pages", func(ctx *fiber.Ctx) error {
		pages, err := service.ListPages(ctx.UserContext())
		if err != nil {
			return mapCatalogError(err)
		}
		return ctx.JSON(pages)
	})
	module.RegisterGET("/api/v1/catalog/pages/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		page, findErr := service.FindPageByID(ctx.UserContext(), id)
		if findErr != nil {
			return mapCatalogError(findErr)
		}
		return ctx.JSON(page)
	})
	module.RegisterPOST("/api/v1/catalog/pages", func(ctx *fiber.Ctx) error {
		var payload domain.CatalogPage
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		created, createErr := service.CreatePage(ctx.UserContext(), payload)
		if createErr != nil {
			return mapCatalogError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(created)
	})
}

// registerOfferRoutes registers catalog offer routes.
func registerOfferRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/catalog/pages/:id/offers", func(ctx *fiber.Ctx) error {
		pageID, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		offers, findErr := service.ListOffersByPageID(ctx.UserContext(), pageID)
		if findErr != nil {
			return mapCatalogError(findErr)
		}
		return ctx.JSON(offers)
	})
}

// registerVoucherRoutes registers voucher redemption routes.
func registerVoucherRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/catalog/vouchers", func(ctx *fiber.Ctx) error {
		vouchers, err := service.ListVouchers(ctx.UserContext())
		if err != nil {
			return mapCatalogError(err)
		}
		return ctx.JSON(vouchers)
	})
	module.RegisterPOST("/api/v1/catalog/vouchers/redeem", func(ctx *fiber.Ctx) error {
		var payload redeemRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		v, redeemErr := service.RedeemVoucher(ctx.UserContext(), payload.Code, payload.UserID)
		if redeemErr != nil {
			return mapCatalogError(redeemErr)
		}
		return ctx.JSON(v)
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

// mapCatalogError maps domain errors to HTTP responses.
func mapCatalogError(err error) error {
	switch err {
	case domain.ErrPageNotFound, domain.ErrOfferNotFound, domain.ErrVoucherNotFound:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrVoucherExhausted, domain.ErrVoucherAlreadyRedeemed:
		return fiber.NewError(http.StatusConflict, err.Error())
	case domain.ErrVoucherDisabled, domain.ErrOfferInactive:
		return fiber.NewError(http.StatusForbidden, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}

// redeemRequest defines voucher redemption payload.
type redeemRequest struct {
	// Code stores voucher code.
	Code string `json:"code"`
	// UserID stores redeeming user identifier.
	UserID int `json:"user_id"`
}
