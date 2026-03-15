package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	permissionapplication "github.com/momlesstomato/pixel-server/pkg/permission/application"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// RegisterRoutes registers permission API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, service Service) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if service == nil {
		return fmt.Errorf("permission service is required")
	}
	registerGroupRoutes(module, service)
	registerAssignmentRoutes(module, service)
	return nil
}

// registerGroupRoutes registers group and permission management routes.
func registerGroupRoutes(module *corehttp.Module, service Service) {
	module.RegisterGET("/api/v1/groups", func(ctx *fiber.Ctx) error {
		groups, err := service.ListGroups(ctx.UserContext())
		if err != nil {
			return mapPermissionError(err)
		}
		return ctx.JSON(fiber.Map{"groups": groups, "count": len(groups)})
	})
	module.RegisterGET("/api/v1/groups/:id", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		group, findErr := service.GetGroup(ctx.UserContext(), groupID)
		if findErr != nil {
			return mapPermissionError(findErr)
		}
		return ctx.JSON(group)
	})
	module.RegisterPOST("/api/v1/groups", func(ctx *fiber.Ctx) error {
		var payload permissionapplication.CreateGroupInput
		if err := ctx.BodyParser(&payload); err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		group, createErr := service.CreateGroup(ctx.UserContext(), payload)
		if createErr != nil {
			return mapPermissionError(createErr)
		}
		return ctx.Status(http.StatusCreated).JSON(group)
	})
	module.RegisterPATCH("/api/v1/groups/:id", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		var payload permissiondomain.GroupPatch
		if bodyErr := ctx.BodyParser(&payload); bodyErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		group, updateErr := service.UpdateGroup(ctx.UserContext(), groupID, payload)
		if updateErr != nil {
			return mapPermissionError(updateErr)
		}
		return ctx.JSON(group)
	})
	module.RegisterDELETE("/api/v1/groups/:id", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if deleteErr := service.DeleteGroup(ctx.UserContext(), groupID); deleteErr != nil {
			return mapPermissionError(deleteErr)
		}
		return ctx.JSON(fiber.Map{"deleted": groupID})
	})
	module.RegisterGET("/api/v1/groups/:id/permissions", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		group, findErr := service.GetGroup(ctx.UserContext(), groupID)
		if findErr != nil {
			return mapPermissionError(findErr)
		}
		return ctx.JSON(fiber.Map{"permissions": group.Permissions, "count": len(group.Permissions)})
	})
	module.RegisterPOST("/api/v1/groups/:id/permissions", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		payload := struct {
			Permissions []string `json:"permissions"`
		}{}
		if bodyErr := ctx.BodyParser(&payload); bodyErr != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid request body")
		}
		group, addErr := service.AddPermissions(ctx.UserContext(), groupID, payload.Permissions)
		if addErr != nil {
			return mapPermissionError(addErr)
		}
		return ctx.JSON(group)
	})
	module.RegisterDELETE("/api/v1/groups/:id/permissions/:permission", func(ctx *fiber.Ctx) error {
		groupID, err := parseIDParam(ctx.Params("id"), "group id")
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		group, removeErr := service.RemovePermission(ctx.UserContext(), groupID, strings.TrimSpace(ctx.Params("permission")))
		if removeErr != nil {
			return mapPermissionError(removeErr)
		}
		return ctx.JSON(group)
	})
}

// parseIDParam parses one positive integer route identifier.
func parseIDParam(value string, label string) (int, error) {
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", label)
	}
	return parsed, nil
}
