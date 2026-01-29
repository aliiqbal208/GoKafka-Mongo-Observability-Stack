package v1

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/middlewares"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	httpErrors "github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/http_errors"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

type cartHandlers struct {
	log      logger.Logger
	cartUC   cart.UseCase
	validate *validator.Validate
	group    fiber.Router
	mw       middlewares.MiddlewareManager
}

// NewCartHandlers creates new cart HTTP handlers
func NewCartHandlers(
	log logger.Logger,
	cartUC cart.UseCase,
	validate *validator.Validate,
	group fiber.Router,
	mw middlewares.MiddlewareManager,
) *cartHandlers {
	return &cartHandlers{
		log:      log,
		cartUC:   cartUC,
		validate: validate,
		group:    group,
		mw:       mw,
	}
}

// GetCart Get user's cart
// @Tags Cart
// @Summary Get user's cart
// @Description Get cart by user ID
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} models.Cart
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cart/{user_id} [get]
func (h *cartHandlers) GetCart() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "cartHandlers.GetCart")
		defer span.Finish()
		getCartRequests.Inc()

		userID := c.Params("user_id")
		if userID == "" {
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		cartDoc, err := h.cartUC.GetCart(ctx, userID)
		if err != nil {
			h.log.Errorf("cartUC.GetCart: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(cartDoc)
	}
}

// AddItem Add item to cart
// @Tags Cart
// @Summary Add item to cart
// @Description Add a product to user's cart
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param item body models.AddItemRequest true "Item to add"
// @Success 200 {object} models.Cart
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cart/{user_id}/items [post]
func (h *cartHandlers) AddItem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "cartHandlers.AddItem")
		defer span.Finish()
		addItemRequests.Inc()

		userID := c.Params("user_id")
		if userID == "" {
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		var req models.AddItemRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("c.BodyParser: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := h.validate.StructCtx(ctx, &req); err != nil {
			h.log.Errorf("validate.StructCtx: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		productID, err := primitive.ObjectIDFromHex(req.ProductID)
		if err != nil {
			h.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		cartDoc, err := h.cartUC.AddItem(ctx, userID, productID, req.Quantity)
		if err != nil {
			h.log.Errorf("cartUC.AddItem: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(cartDoc)
	}
}

// UpdateItem Update item quantity
// @Tags Cart
// @Summary Update cart item quantity
// @Description Update the quantity of an item in cart
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param product_id path string true "Product ID"
// @Param item body models.UpdateItemRequest true "New quantity"
// @Success 200 {object} models.Cart
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cart/{user_id}/items/{product_id} [put]
func (h *cartHandlers) UpdateItem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "cartHandlers.UpdateItem")
		defer span.Finish()
		updateItemRequests.Inc()

		userID := c.Params("user_id")
		productIDStr := c.Params("product_id")

		if userID == "" || productIDStr == "" {
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		productID, err := primitive.ObjectIDFromHex(productIDStr)
		if err != nil {
			h.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		var req models.UpdateItemRequest
		if err := c.BodyParser(&req); err != nil {
			h.log.Errorf("c.BodyParser: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := h.validate.StructCtx(ctx, &req); err != nil {
			h.log.Errorf("validate.StructCtx: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		cartDoc, err := h.cartUC.UpdateItemQuantity(ctx, userID, productID, req.Quantity)
		if err != nil {
			h.log.Errorf("cartUC.UpdateItemQuantity: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(cartDoc)
	}
}

// RemoveItem Remove item from cart
// @Tags Cart
// @Summary Remove item from cart
// @Description Remove a product from user's cart
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param product_id path string true "Product ID"
// @Success 200 {object} models.Cart
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cart/{user_id}/items/{product_id} [delete]
func (h *cartHandlers) RemoveItem() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "cartHandlers.RemoveItem")
		defer span.Finish()
		removeItemRequests.Inc()

		userID := c.Params("user_id")
		productIDStr := c.Params("product_id")

		if userID == "" || productIDStr == "" {
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		productID, err := primitive.ObjectIDFromHex(productIDStr)
		if err != nil {
			h.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		cartDoc, err := h.cartUC.RemoveItem(ctx, userID, productID)
		if err != nil {
			h.log.Errorf("cartUC.RemoveItem: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(cartDoc)
	}
}

// ClearCart Clear all items from cart
// @Tags Cart
// @Summary Clear cart
// @Description Remove all items from user's cart
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cart/{user_id} [delete]
func (h *cartHandlers) ClearCart() fiber.Handler {
	return func(c *fiber.Ctx) error {
		span, ctx := opentracing.StartSpanFromContext(context.Background(), "cartHandlers.ClearCart")
		defer span.Finish()
		clearCartRequests.Inc()

		userID := c.Params("user_id")
		if userID == "" {
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, httpErrors.BadRequest)
		}

		if err := h.cartUC.ClearCart(ctx, userID); err != nil {
			h.log.Errorf("cartUC.ClearCart: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "Cart cleared successfully",
		})
	}
}
