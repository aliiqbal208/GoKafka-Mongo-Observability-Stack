package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/config"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

var (
	httpTotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "http_microservice_total_requests",
		Help: "The total number of incoming HTTP requests",
	})
)

// MiddlewareManager http middlewares
type middlewareManager struct {
	log logger.Logger
	cfg *config.Config
}

// MiddlewareManager interface
type MiddlewareManager interface {
	Metrics(c *fiber.Ctx) error
}

// NewMiddlewareManager constructor
func NewMiddlewareManager(log logger.Logger, cfg *config.Config) *middlewareManager {
	return &middlewareManager{log: log, cfg: cfg}
}

// Metrics prometheus metrics
func (m *middlewareManager) Metrics(c *fiber.Ctx) error {
	httpTotalRequests.Inc()
	return c.Next()
}
