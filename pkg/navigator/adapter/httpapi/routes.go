package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/navigator/domain"
)

// RegisterRoutes registers navigator API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("navigator service is required")
	}
	registerCategoryRoutes(module, service)
	registerRoomRoutes(module, service)
	return nil
}

// registerCategoryRoutes registers navigator category CRUD routes.
func registerCategoryRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/navigator/categories", func(ctx *fiber.Ctx) error {
		cats, err := service.ListCategories(ctx.UserContext())
		if err != nil {
			return mapNavError(err)
		}
		return ctx.JSON(cats)
	})
	module.RegisterPOST("/api/v1/navigator/categories", func(ctx *fiber.Ctx) error {
		var payload domain.Category
		if parseErr := ctx.BodyParser(&payload); parseErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		created, createErr := service.CreateCategory(ctx.UserContext(), payload)
		if createErr != nil {
			return mapNavError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(created)
	})
	module.RegisterDELETE("/api/v1/navigator/categories/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if deleteErr := service.DeleteCategory(ctx.UserContext(), id); deleteErr != nil {
			return mapNavError(deleteErr)
		}
		return ctx.SendStatus(http.StatusNoContent)
	})
}

// registerRoomRoutes registers room query and management routes.
func registerRoomRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/navigator/rooms", func(ctx *fiber.Ctx) error {
		filter := domain.RoomFilter{
			SearchQuery: ctx.Query("q"),
			Offset:      ctx.QueryInt("offset"),
			Limit:       ctx.QueryInt("limit", 20),
		}
		rooms, total, err := service.ListRooms(ctx.UserContext(), filter)
		if err != nil {
			return mapNavError(err)
		}
		return ctx.JSON(fiber.Map{"rooms": rooms, "total": total})
	})
	module.RegisterGET("/api/v1/navigator/rooms/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		room, findErr := service.FindRoomByID(ctx.UserContext(), id)
		if findErr != nil {
			return mapNavError(findErr)
		}
		return ctx.JSON(room)
	})
	module.RegisterDELETE("/api/v1/navigator/rooms/:id", func(ctx *fiber.Ctx) error {
		id, err := parsePositiveID(ctx.Params("id"))
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if deleteErr := service.DeleteRoom(ctx.UserContext(), id); deleteErr != nil {
			return mapNavError(deleteErr)
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

// mapNavError maps domain errors to HTTP responses.
func mapNavError(err error) error {
	switch err {
	case domain.ErrCategoryNotFound, domain.ErrRoomNotFound:
		return fiber.NewError(http.StatusNotFound, err.Error())
	case domain.ErrFavouriteLimitReached, domain.ErrFavouriteAlreadyExists:
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "must") || strings.Contains(msg, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
