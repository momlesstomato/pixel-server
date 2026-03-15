package httpapi

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	permissiondomain "github.com/momlesstomato/pixel-server/pkg/permission/domain"
)

// mapPermissionError maps permission domain/application errors to HTTP responses.
func mapPermissionError(err error) error {
	if err == nil {
		return nil
	}
	if err == permissiondomain.ErrGroupNotFound {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	if err == permissiondomain.ErrGroupInUse || err == permissiondomain.ErrCannotDeleteDefaultGroup || err == permissiondomain.ErrDefaultGroupRequired {
		return fiber.NewError(http.StatusConflict, err.Error())
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "must") || strings.Contains(message, "invalid") || strings.Contains(message, "required") {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	return fiber.NewError(http.StatusInternalServerError, err.Error())
}
