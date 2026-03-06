package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"pixelsv/internal/auth/domain"
	httpserver "pixelsv/pkg/http"
)

// TicketService defines auth application behavior required by HTTP routes.
type TicketService interface {
	// CreateTicket creates one user ticket with optional ttl seconds.
	CreateTicket(userID int32, ttlSeconds int32) (string, int32, error)
	// RevokeTicket revokes one ticket by value.
	RevokeTicket(ticket string) error
}

// CreateTicketRequest defines ticket creation payload.
type CreateTicketRequest struct {
	// UserID is the target authenticated user id.
	UserID int32 `json:"user_id"`
	// TTLSeconds is optional ticket validity duration in seconds.
	TTLSeconds int32 `json:"ttl_seconds"`
}

// CreateTicketResponse defines ticket creation response payload.
type CreateTicketResponse struct {
	// Ticket is the generated one-time ticket string.
	Ticket string `json:"ticket"`
	// UserID is the associated user id.
	UserID int32 `json:"user_id"`
	// TTLSeconds is the applied ticket ttl in seconds.
	TTLSeconds int32 `json:"ttl_seconds"`
}

// RegisterRoutes mounts auth admin routes on the provided app.
func RegisterRoutes(app *fiber.App, service TicketService, apiKey string) {
	group := app.Group("/api/v1/auth", httpserver.APIKeyMiddleware(apiKey))
	group.Post("/tickets", func(c *fiber.Ctx) error {
		request := CreateTicketRequest{}
		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
		}
		ticket, ttlSeconds, err := service.CreateTicket(request.UserID, request.TTLSeconds)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidUserID) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ticket creation failed"})
		}
		response := CreateTicketResponse{Ticket: ticket, UserID: request.UserID, TTLSeconds: ttlSeconds}
		return c.Status(fiber.StatusCreated).JSON(response)
	})
	group.Delete("/tickets/:ticket", func(c *fiber.Ctx) error {
		ticket := c.Params("ticket")
		if err := service.RevokeTicket(ticket); err != nil {
			if errors.Is(err, domain.ErrInvalidTicket) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ticket revoke failed"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})
}
