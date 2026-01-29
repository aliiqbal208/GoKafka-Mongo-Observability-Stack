package v1

import (
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/middlewares"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/user"
	"github.com/gofiber/fiber/v2"
)

// MapAuthRoutes maps auth routes
func MapAuthRoutes(router fiber.Router, h user.Handlers, mw middlewares.MiddlewareManager) {
	authGroup := router.Group("/auth")
	authGroup.Post("/signup", h.Signup())
	authGroup.Post("/login", h.Login())
	authGroup.Post("/logout", h.Logout())
	authGroup.Get("/me", mw.RequireAuth(), h.GetCurrentUser())
}
