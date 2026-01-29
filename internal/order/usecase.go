package order

import (
	"context"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
)

// UseCase defines the order use case interface
type UseCase interface {
	CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error)
	GetOrderByID(ctx context.Context, orderID string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID string, page, size int64) (*models.OrdersList, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error)
	CancelOrder(ctx context.Context, orderID string) (*models.Order, error)
}
