package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/docs"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// runHTTPServer starts the HTTP server. This call will BLOCK until the server is closed or errors out.
func (s *server) runHTTPServer() error {
	// Basic routes
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server running on HTTP")
	})

	s.app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Test route working!")
	})

	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	s.app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Configure Swagger *for HTTP* (no HTTPS).
	docs.SwaggerInfo.Title = "Products microservice"
	docs.SwaggerInfo.Description = "Products REST API microservice."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = "localhost:5007"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// Swagger routes - serve static swagger UI
	s.app.Get("/swagger/*", func(c *fiber.Ctx) error {
		// For swagger, we'll redirect to the swagger docs
		return c.Redirect("/swagger/index.html", http.StatusMovedPermanently)
	})

	// Common middleware:
	s.app.Use(logger.New())
	s.app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, X-Request-ID, " + csrfTokenHeader,
	}))
	s.app.Use(recover.New())
	s.app.Use(requestid.New())
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
		Next: func(c *fiber.Ctx) bool {
			return strings.Contains(c.Path(), "swagger")
		},
	}))

	addr := s.cfg.Http.Port
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	s.app.Server().ReadTimeout = time.Second * s.cfg.Http.ReadTimeout
	s.app.Server().WriteTimeout = time.Second * s.cfg.Http.WriteTimeout
	s.app.Server().MaxRequestBodySize = maxHeaderBytes

	s.log.Infof("Starting HTTP server on %s", addr)
	// BLOCKING call (will not return until the server is closed or fails).
	return s.app.Listen(addr)
}
