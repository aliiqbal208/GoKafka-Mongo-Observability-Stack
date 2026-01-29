package order

import (
	"context"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
)

// MongoRepository defines the order MongoDB repository interface
type MongoRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, error)
	GetByID(ctx context.Context, orderID string) (*models.Order, error)
	GetByUserID(ctx context.Context, userID string, page, size int64) (*models.OrdersList, error)
	UpdateStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) (*models.Order, error)
	Delete(ctx context.Context, orderID string) error
}

// RedisRepository defines the order Redis cache repository interface
type RedisRepository interface {
	GetByID(ctx context.Context, orderID string) (*models.Order, error)
	SetOrder(ctx context.Context, order *models.Order) error
	DeleteOrder(ctx context.Context, orderID string) error
}
