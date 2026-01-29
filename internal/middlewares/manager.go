package middlewares

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/jwt"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

var (
	httpTotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "http_microservice_total_requests",
		Help: "The total number of incoming HTTP requests",
	})
)

const (
	UserIDKey    = "userId"
	UserEmailKey = "userEmail"
	UserNameKey  = "userName"
)

// MiddlewareManager http middlewares
type middlewareManager struct {
	log        logger.Logger
	cfg        *config.Config
	jwtManager *jwt.Manager
}

// MiddlewareManager interface
type MiddlewareManager interface {
	Metrics(c *fiber.Ctx) error
	RequireAuth() fiber.Handler
}

// NewMiddlewareManager constructor
func NewMiddlewareManager(log logger.Logger, cfg *config.Config, jwtManager *jwt.Manager) *middlewareManager {
	return &middlewareManager{log: log, cfg: cfg, jwtManager: jwtManager}
}

// Metrics prometheus metrics
func (m *middlewareManager) Metrics(c *fiber.Ctx) error {
	httpTotalRequests.Inc()
	return c.Next()
}

// RequireAuth middleware checks if user is authenticated via JWT
func (m *middlewareManager) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format. Use: Bearer <token>",
			})
		}

		tokenString := parts[1]
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			m.log.Errorf("RequireAuth.ValidateToken: %v", err)
			if err == jwt.ErrExpiredToken {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token has expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Store user info in context for downstream handlers
		c.Locals(UserIDKey, claims.UserID)
		c.Locals(UserEmailKey, claims.Email)
		c.Locals(UserNameKey, claims.Name)

		return c.Next()
	}
}
