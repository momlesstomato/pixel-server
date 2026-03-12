package httpapi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	corehttp "github.com/momlesstomato/pixel-server/core/http"
	"github.com/momlesstomato/pixel-server/pkg/authentication"
)

// Issuer defines SSO ticket issuance behavior required by HTTP routes.
type Issuer interface {
	// Issue generates and stores one SSO ticket.
	Issue(context.Context, authentication.IssueRequest) (authentication.IssueResult, error)
}

// RegisterRoutes registers authentication API routes on an HTTP module.
func RegisterRoutes(module *corehttp.Module, issuer Issuer) error {
	if module == nil {
		return fmt.Errorf("http module is required")
	}
	if issuer == nil {
		return fmt.Errorf("issuer is required")
	}
	module.RegisterPOST("/api/v1/sso", func(ctx *fiber.Ctx) error {
		var payload issueRequest
		if err := ctx.BodyParser(&payload); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		ttl := time.Duration(payload.TTLSeconds) * time.Second
		result, err := issuer.Issue(ctx.UserContext(), authentication.IssueRequest{
			UserID: payload.UserID, TTL: ttl,
		})
		if err != nil {
			if strings.Contains(err.Error(), "user id") || strings.Contains(err.Error(), "ttl") {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return ctx.JSON(issueResponse{
			Ticket: result.Ticket, ExpiresAt: result.ExpiresAt.UTC().Format(time.RFC3339),
		})
	})
	return nil
}

// issueRequest defines ticket issuance request payload.
type issueRequest struct {
	// UserID defines user identifier used for SSO ticket binding.
	UserID int `json:"user_id"`
	// TTLSeconds defines optional ticket lifetime override in seconds.
	TTLSeconds int64 `json:"ttl_seconds"`
}

// issueResponse defines ticket issuance response payload.
type issueResponse struct {
	// Ticket defines issued single-use token.
	Ticket string `json:"ticket"`
	// ExpiresAt defines ticket expiration time in RFC3339 format.
	ExpiresAt string `json:"expires_at"`
}
