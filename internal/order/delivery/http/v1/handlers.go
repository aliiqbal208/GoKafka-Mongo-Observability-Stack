package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/order"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

type orderHandlers struct {
	log   logger.Logger
	uc    order.UseCase
	group fiber.Router
}

// NewOrderHandlers creates new order handlers
func NewOrderHandlers(log logger.Logger, uc order.UseCase, group fiber.Router) *orderHandlers {
	return &orderHandlers{log: log, uc: uc, group: group}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Creates a new order from the user's cart
// @Tags Orders
// @Accept json
// @Produce json
// @Param body body models.CreateOrderRequest true "Order creation request"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [post]
func (h *orderHandlers) CreateOrder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.Context(), "orderHandlers.CreateOrder")
		defer span.Finish()

		var req models.CreateOrderRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("orderHandlers.CreateOrder.BodyParser: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if req.UserID == "" || req.PaymentMethod == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "user_id and payment_method are required",
			})
		}

		createdOrder, err := h.uc.CreateOrder(ctx, &req)
		if err != nil {
			h.log.Errorf("orderHandlers.CreateOrder.uc.CreateOrder: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(createdOrder)
	}
}

// GetOrder godoc
// @Summary Get an order by ID
// @Description Retrieves an order by its ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{order_id} [get]
func (h *orderHandlers) GetOrder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.Context(), "orderHandlers.GetOrder")
		defer span.Finish()

		orderID := c.Params("order_id")
		if orderID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "order_id is required",
			})
		}

		orderDoc, err := h.uc.GetOrderByID(ctx, orderID)
		if err != nil {
			h.log.Errorf("orderHandlers.GetOrder.uc.GetOrderByID: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(orderDoc)
	}
}

// GetUserOrders godoc
// @Summary Get all orders for a user
// @Description Retrieves all orders for a specific user with pagination
// @Tags Orders
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param size query int false "Page size" default(10)
// @Success 200 {object} models.OrdersList
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/user/{user_id} [get]
func (h *orderHandlers) GetUserOrders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.Context(), "orderHandlers.GetUserOrders")
		defer span.Finish()

		userID := c.Params("user_id")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "user_id is required",
			})
		}

		page, _ := strconv.ParseInt(c.Query("page", "1"), 10, 64)
		size, _ := strconv.ParseInt(c.Query("size", "10"), 10, 64)

		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 10
		}

		ordersList, err := h.uc.GetUserOrders(ctx, userID, page, size)
		if err != nil {
			h.log.Errorf("orderHandlers.GetUserOrders.uc.GetUserOrders: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(ordersList)
	}
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Updates the status of an order
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order ID"
// @Param body body models.UpdateOrderStatusRequest true "Status update request"
// @Success 200 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{order_id}/status [put]
func (h *orderHandlers) UpdateOrderStatus() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.Context(), "orderHandlers.UpdateOrderStatus")
		defer span.Finish()

		orderID := c.Params("order_id")
		if orderID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "order_id is required",
			})
		}

		var req models.UpdateOrderStatusRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("orderHandlers.UpdateOrderStatus.BodyParser: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid request body",
			})
		}

		if !req.Status.IsValid() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid status",
			})
		}

		updatedOrder, err := h.uc.UpdateOrderStatus(ctx, orderID, req.Status)
		if err != nil {
			h.log.Errorf("orderHandlers.UpdateOrderStatus.uc.UpdateOrderStatus: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(updatedOrder)
	}
}

// CancelOrder godoc
// @Summary Cancel an order
// @Description Cancels an order (only for pending or confirmed orders)
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders/{order_id}/cancel [put]
func (h *orderHandlers) CancelOrder() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(c.Context(), "orderHandlers.CancelOrder")
		defer span.Finish()

		orderID := c.Params("order_id")
		if orderID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "order_id is required",
			})
		}

		cancelledOrder, err := h.uc.CancelOrder(ctx, orderID)
		if err != nil {
			h.log.Errorf("orderHandlers.CancelOrder.uc.CancelOrder: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		return c.Status(fiber.StatusOK).JSON(cancelledOrder)
	}
}
