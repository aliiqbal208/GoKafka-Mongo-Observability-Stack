package usecase

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/product"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

// cartUseCase implements cart.UseCase
type cartUseCase struct {
	log         logger.Logger
	mongoRepo   cart.MongoRepository
	redisRepo   cart.RedisRepository
	productRepo product.MongoRepository
}

// NewCartUseCase creates a new cart use case
func NewCartUseCase(
	log logger.Logger,
	mongoRepo cart.MongoRepository,
	redisRepo cart.RedisRepository,
	productRepo product.MongoRepository,
) cart.UseCase {
	return &cartUseCase{
		log:         log,
		mongoRepo:   mongoRepo,
		redisRepo:   redisRepo,
		productRepo: productRepo,
	}
}

// GetCart retrieves a user's cart
func (uc *cartUseCase) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartUseCase.GetCart")
	defer span.Finish()

	// Try to get from cache first
	cachedCart, err := uc.redisRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.log.Warnf("Redis cache error: %v", err)
	}

	if cachedCart != nil {
		return cachedCart, nil
	}

	// Get from MongoDB
	cartDoc, err := uc.mongoRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.GetByUserID")
	}

	// If no cart exists, return empty cart
	if cartDoc == nil {
		return &models.Cart{
			UserID:    userID,
			Items:     []models.CartItem{},
			Total:     0,
			ItemCount: 0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	// Cache the cart
	if err := uc.redisRepo.SetCart(ctx, cartDoc); err != nil {
		uc.log.Warnf("Failed to cache cart: %v", err)
	}

	return cartDoc, nil
}

// AddItem adds an item to the cart
func (uc *cartUseCase) AddItem(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartUseCase.AddItem")
	defer span.Finish()

	// Validate product exists and get its details
	productDoc, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "productRepo.GetByID")
	}

	if productDoc == nil {
		return nil, errors.New("product not found")
	}

	// Check stock availability
	if productDoc.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// Create cart item
	cartItem := models.CartItem{
		ProductID:   productDoc.ProductID,
		ProductName: productDoc.Name,
		Price:       productDoc.Price,
		Quantity:    quantity,
		ImageURL:    productDoc.GetImage(),
	}

	// Add item to cart
	updatedCart, err := uc.mongoRepo.AddItem(ctx, userID, cartItem)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.AddItem")
	}

	// Invalidate cache
	if err := uc.redisRepo.DeleteCart(ctx, userID); err != nil {
		uc.log.Warnf("Failed to invalidate cart cache: %v", err)
	}

	// Cache the updated cart
	if err := uc.redisRepo.SetCart(ctx, updatedCart); err != nil {
		uc.log.Warnf("Failed to cache cart: %v", err)
	}

	return updatedCart, nil
}

// UpdateItemQuantity updates the quantity of an item
func (uc *cartUseCase) UpdateItemQuantity(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartUseCase.UpdateItemQuantity")
	defer span.Finish()

	if quantity <= 0 {
		return uc.RemoveItem(ctx, userID, productID)
	}

	// Validate stock availability
	productDoc, err := uc.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "productRepo.GetByID")
	}

	if productDoc == nil {
		return nil, errors.New("product not found")
	}

	if productDoc.Stock < quantity {
		return nil, errors.New("insufficient stock")
	}

	// Update item quantity
	updatedCart, err := uc.mongoRepo.UpdateItemQuantity(ctx, userID, productID, quantity)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.UpdateItemQuantity")
	}

	// Invalidate and update cache
	if err := uc.redisRepo.DeleteCart(ctx, userID); err != nil {
		uc.log.Warnf("Failed to invalidate cart cache: %v", err)
	}

	if err := uc.redisRepo.SetCart(ctx, updatedCart); err != nil {
		uc.log.Warnf("Failed to cache cart: %v", err)
	}

	return updatedCart, nil
}

// RemoveItem removes an item from the cart
func (uc *cartUseCase) RemoveItem(ctx context.Context, userID string, productID primitive.ObjectID) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartUseCase.RemoveItem")
	defer span.Finish()

	updatedCart, err := uc.mongoRepo.RemoveItem(ctx, userID, productID)
	if err != nil {
		return nil, errors.Wrap(err, "mongoRepo.RemoveItem")
	}

	// Invalidate and update cache
	if err := uc.redisRepo.DeleteCart(ctx, userID); err != nil {
		uc.log.Warnf("Failed to invalidate cart cache: %v", err)
	}

	if err := uc.redisRepo.SetCart(ctx, updatedCart); err != nil {
		uc.log.Warnf("Failed to cache cart: %v", err)
	}

	return updatedCart, nil
}

// ClearCart removes all items from the cart
func (uc *cartUseCase) ClearCart(ctx context.Context, userID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartUseCase.ClearCart")
	defer span.Finish()

	if err := uc.mongoRepo.Clear(ctx, userID); err != nil {
		return errors.Wrap(err, "mongoRepo.Clear")
	}

	// Invalidate cache
	if err := uc.redisRepo.DeleteCart(ctx, userID); err != nil {
		uc.log.Warnf("Failed to invalidate cart cache: %v", err)
	}

	return nil
}
