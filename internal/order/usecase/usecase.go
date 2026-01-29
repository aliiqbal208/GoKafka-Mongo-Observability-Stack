package usecase

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

// orderUseCase implements order.UseCase
type orderUseCase struct {
	log         logger.Logger
	mongoRepo   order.MongoRepository
	redisRepo   order.RedisRepository
	cartUC      cart.UseCase
	productRepo product.MongoRepository
}

// NewOrderUseCase creates a new order use case
func NewOrderUseCase(
	log logger.Logger,
	mongoRepo order.MongoRepository,
	redisRepo order.RedisRepository,
	cartUC cart.UseCase,
	productRepo product.MongoRepository,
) order.UseCase {
	return &orderUseCase{
		log:         log,
		mongoRepo:   mongoRepo,
		redisRepo:   redisRepo,
		cartUC:      cartUC,
		productRepo: productRepo,
	}
}

// CreateOrder creates a new order from the user's cart
func (uc *orderUseCase) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderUseCase.CreateOrder")
	defer span.Finish()

	// Get the user's cart
	userCart, err := uc.cartUC.GetCart(ctx, req.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "cartUC.GetCart")
	}

	if userCart == nil || len(userCart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Validate stock and reserve inventory
	for _, item := range userCart.Items {
		productDoc, err := uc.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, errors.Wrap(err, "productRepo.GetByID")
		}

		if productDoc == nil {
			return nil, errors.Errorf("product %s not found", item.ProductID.Hex())
		}

		if productDoc.Stock < item.Quantity {
			return nil, errors.Errorf("insufficient stock for product %s", productDoc.Name)
		}
	}

	// Convert cart items to order items
	orderItems := userCart.ToOrderItems()

	// Create the order
	newOrder := &models.Order{
		UserID:          req.UserID,
		Items:           orderItems,
		TotalAmount:     userCart.Total,
		Status:          models.OrderStatusPending,
		ShippingAddress: req.ShippingAddress,
		PaymentMethod:   req.PaymentMethod,
		Notes:           req.Notes,
	}

	createdOrder, err := uc.mongoRepo.Create(ctx, newOrder)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.Create")
	}

	// Clear the user's cart after successful order creation
	if err := uc.cartUC.ClearCart(ctx, req.UserID); err != nil {
		uc.log.Warnf("Failed to clear cart after order creation: %v", err)
	}

	// Cache the order
	if err := uc.redisRepo.SetOrder(ctx, createdOrder); err != nil {
		uc.log.Warnf("Failed to cache order: %v", err)
	}

	return createdOrder, nil
}

// GetOrderByID retrieves an order by its ID
func (uc *orderUseCase) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderUseCase.GetOrderByID")
	defer span.Finish()

	// Try cache first
	cachedOrder, err := uc.redisRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.log.Warnf("Redis cache error: %v", err)
	}

	if cachedOrder != nil {
		return cachedOrder, nil
	}

	// Get from MongoDB
	orderDoc, err := uc.mongoRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.GetByID")
	}

	if orderDoc == nil {
		return nil, errors.New("order not found")
	}

	// Cache the order
	if err := uc.redisRepo.SetOrder(ctx, orderDoc); err != nil {
		uc.log.Warnf("Failed to cache order: %v", err)
	}

	return orderDoc, nil
}

// GetUserOrders retrieves all orders for a user with pagination
func (uc *orderUseCase) GetUserOrders(ctx context.Context, userID string, page, size int64) (*models.OrdersList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderUseCase.GetUserOrders")
	defer span.Finish()

	ordersList, err := uc.mongoRepo.GetByUserID(ctx, userID, page, size)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.GetByUserID")
	}

	return ordersList, nil
}

// UpdateOrderStatus updates the status of an order
func (uc *orderUseCase) UpdateOrderStatus(ctx context.Context, orderID string, status models.OrderStatus) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderUseCase.UpdateOrderStatus")
	defer span.Finish()

	// Validate status transition
	existingOrder, err := uc.mongoRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.GetByID")
	}

	if existingOrder == nil {
		return nil, errors.New("order not found")
	}

	// Check if status transition is valid
	if !isValidStatusTransition(existingOrder.Status, status) {
		return nil, errors.Errorf("invalid status transition from %s to %s", existingOrder.Status, status)
	}

	updatedOrder, err := uc.mongoRepo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.UpdateStatus")
	}

	// Invalidate cache
	if err := uc.redisRepo.DeleteOrder(ctx, orderID); err != nil {
		uc.log.Warnf("Failed to invalidate order cache: %v", err)
	}

	// Cache updated order
	if err := uc.redisRepo.SetOrder(ctx, updatedOrder); err != nil {
		uc.log.Warnf("Failed to cache updated order: %v", err)
	}

	return updatedOrder, nil
}

// CancelOrder cancels an order
func (uc *orderUseCase) CancelOrder(ctx context.Context, orderID string) (*models.Order, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "orderUseCase.CancelOrder")
	defer span.Finish()

	existingOrder, err := uc.mongoRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.GetByID")
	}

	if existingOrder == nil {
		return nil, errors.New("order not found")
	}

	// Can only cancel pending or confirmed orders
	if existingOrder.Status != models.OrderStatusPending && existingOrder.Status != models.OrderStatusConfirmed {
		return nil, errors.Errorf("cannot cancel order with status %s", existingOrder.Status)
	}

	return uc.UpdateOrderStatus(ctx, orderID, models.OrderStatusCancelled)
}

// isValidStatusTransition checks if a status transition is valid
func isValidStatusTransition(from, to models.OrderStatus) bool {
	validTransitions := map[models.OrderStatus][]models.OrderStatus{
		models.OrderStatusPending:    {models.OrderStatusConfirmed, models.OrderStatusCancelled},
		models.OrderStatusConfirmed:  {models.OrderStatusProcessing, models.OrderStatusCancelled},
		models.OrderStatusProcessing: {models.OrderStatusShipped},
		models.OrderStatusShipped:    {models.OrderStatusDelivered},
		models.OrderStatusDelivered:  {},
		models.OrderStatusCancelled:  {},
	}

	allowedTransitions, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowed := range allowedTransitions {
		if allowed == to {
			return true
		}
	}

	return false
}
