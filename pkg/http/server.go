package httpserver

import (
	"context"
	"time"

	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Server owns the core Fiber HTTP/WebSocket runtime.
type Server struct {
	app    *fiber.App
	cfg    Config
	logger *zap.Logger
}

// New creates a new Server instance.
func New(cfg Config, logger *zap.Logger) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	app := fiber.New(fiber.Config{
		DisableStartupMessage: cfg.DisableStartupMessage,
		ReadTimeout:           time.Duration(cfg.ReadTimeoutSeconds) * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error("fiber request error", zap.String("path", c.Path()), zap.Error(err))
			return fiber.DefaultErrorHandler(c, err)
		},
	})
	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
		Next: func(*fiber.Ctx) bool {
			return !logger.Core().Enabled(zap.DebugLevel)
		},
	}))
	server := &Server{app: app, cfg: cfg, logger: logger}
	server.registerRoutes()
	return server, nil
}

// App returns the underlying Fiber app.
func (s *Server) App() *fiber.App {
	return s.app
}

// ListenAndServe starts the server and shuts it down when context is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	s.logger.Info("http server listening", zap.String("address", s.cfg.Address))
	go func() {
		<-ctx.Done()
		s.logger.Info("http server shutdown requested")
		_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.app.Shutdown()
	}()
	return s.app.Listen(s.cfg.Address)
}

func (s *Server) registerRoutes() {
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
	s.app.Get("/ready", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ready"})
	})
	s.app.Get(s.cfg.OpenAPIPath, func(c *fiber.Ctx) error {
		c.Type("application/json")
		return c.Send(OpenAPISpecJSON())
	})
	s.app.Get(s.cfg.SwaggerPath, func(c *fiber.Ctx) error {
		c.Type("html")
		return c.SendString(SwaggerHTML(s.cfg.OpenAPIPath))
	})
	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return c.SendStatus(fiber.StatusUpgradeRequired)
	})
	s.app.Get("/ws", websocket.New(func(conn *websocket.Conn) {
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}))
	admin := s.app.Group("/api/v1", APIKeyMiddleware(s.cfg.APIKey))
	admin.Get("/admin/ping", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok", "scope": "admin"})
	})
}
