package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// LocalKeyRayID defines the fiber context local key for the request ray identifier.
const LocalKeyRayID = "ray_id"

// HeaderRayID defines the HTTP response header that carries the request trace identifier.
const HeaderRayID = "X-Ray-ID"

// TraceMiddleware returns a Fiber handler that generates a UUID ray_id per request,
// stores it in context locals, and attaches it to the response header.
func TraceMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := uuid.New().String()
		ctx.Locals(LocalKeyRayID, id)
		ctx.Set(HeaderRayID, id)
		return ctx.Next()
	}
}

// RayID extracts the ray_id string from a Fiber request context.
func RayID(ctx *fiber.Ctx) string {
	id, _ := ctx.Locals(LocalKeyRayID).(string)
	return id
}
