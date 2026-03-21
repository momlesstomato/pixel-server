package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/furniture/domain"
)

// RegisterRoutes registers furniture API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("furniture service is required")
	}
	registerDefinitionRoutes(module, service)
	registerItemRoutes(module, service)
	return nil
}

// registerDefinitionRoutes registers item definition CRUD routes.
func registerDefinitionRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/furniture/definitions", func(ctx *fiber.Ctx) error {
		defs, err := service.ListDefinitions(ctx.UserContext())
		if err != nil {
			return mapFurnitureError(err)
		}
		return ctx.JSON(defs)
	})
	module.RegisterGET("/api/v1/furniture/definitions/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		def, findErr := service.FindDefinitionByID(ctx.UserContext(), id)
		if findErr != nil {
			return mapFurnitureError(findErr)
		}
		return ctx.JSON(def)
	})
	module.RegisterPOST("/api/v1/furniture/definitions", func(ctx *fiber.Ctx) error {
		var payload domain.Definition
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		created, createErr := service.CreateDefinition(ctx.UserContext(), payload)
		if createErr != nil {
			return mapFurnitureError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(created)
	})
	module.RegisterPATCH("/api/v1/furniture/definitions/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var patch domain.DefinitionPatch
		if parseErr := ctx.BodyParser(&patch); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		updated, updateErr := service.UpdateDefinition(ctx.UserContext(), id, patch)
		if updateErr != nil {
			return mapFurnitureError(updateErr)
		}
		return ctx.JSON(updated)
	})
	module.RegisterDELETE("/api/v1/furniture/definitions/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if deleteErr := service.DeleteDefinition(ctx.UserContext(), id); deleteErr != nil {
			return mapFurnitureError(deleteErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
}

// registerItemRoutes registers item instance routes.
func registerItemRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/furniture/items/user/:userId", func(ctx *fiber.Ctx) error {
		userID, err := parsePositiveID(ctx.Params("userId"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		items, findErr := service.ListItemsByUserID(ctx.UserContext(), userID)
		if findErr != nil {
			return mapFurnitureError(findErr)
		}
		return ctx.JSON(items)
	})
	module.RegisterPOST("/api/v1/furniture/items/:id/transfer", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload transferRequest
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		if transferErr := service.TransferItem(ctx.UserContext(), id, payload.NewUserID); transferErr != nil {
			return mapFurnitureError(transferErr)
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

// mapFurnitureError maps domain errors to HTTP responses.
func mapFurnitureError(err error) error {
	if err == domain.ErrDefinitionNotFound || err == domain.ErrItemNotFound {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	if err == domain.ErrItemNotOwned || err == domain.ErrItemNotTradable {
		return fiber.NewError(http.StatusForbidden, err.Error())
	}
	if err == domain.ErrLimitedSoldOut {
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}

// transferRequest defines item transfer payload.
type transferRequest struct {
	// NewUserID stores new owner identifier.
	NewUserID int `json:"new_user_id"`
}
