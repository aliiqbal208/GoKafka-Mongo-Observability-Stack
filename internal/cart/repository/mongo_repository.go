package repository

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/cart"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/pkg/logger"
)

const (
	cartsCollection = "carts"
)

// cartMongoRepository implements cart.MongoRepository
type cartMongoRepository struct {
	log logger.Logger
	db  *mongo.Database
}

// NewCartMongoRepository creates a new cart MongoDB repository
func NewCartMongoRepository(log logger.Logger, db *mongo.Database) cart.MongoRepository {
	return &cartMongoRepository{log: log, db: db}
}

// GetByUserID retrieves a cart by user ID
func (r *cartMongoRepository) GetByUserID(ctx context.Context, userID string) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.GetByUserID")
	defer span.Finish()

	var cartDoc models.Cart
	err := r.db.Collection(cartsCollection).FindOne(ctx, bson.M{"userId": userID}).Decode(&cartDoc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Cart doesn't exist yet
		}
		r.log.Errorf("GetByUserID error: %v", err)
		return nil, errors.Wrap(err, "db.Collection.FindOne")
	}

	return &cartDoc, nil
}

// Create creates a new cart
func (r *cartMongoRepository) Create(ctx context.Context, cart *models.Cart) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.Create")
	defer span.Finish()

	cart.CartID = primitive.NewObjectID()
	cart.CreatedAt = time.Now()
	cart.UpdatedAt = time.Now()
	cart.CalculateTotal()

	_, err := r.db.Collection(cartsCollection).InsertOne(ctx, cart)
	if err != nil {
		r.log.Errorf("Create error: %v", err)
		return nil, errors.Wrap(err, "db.Collection.InsertOne")
	}

	return cart, nil
}

// Update updates an existing cart
func (r *cartMongoRepository) Update(ctx context.Context, cart *models.Cart) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.Update")
	defer span.Finish()

	cart.UpdatedAt = time.Now()
	cart.CalculateTotal()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedCart models.Cart
	err := r.db.Collection(cartsCollection).FindOneAndUpdate(
		ctx,
		bson.M{"userId": cart.UserID},
		bson.M{"$set": cart},
		opts,
	).Decode(&updatedCart)

	if err != nil {
		r.log.Errorf("Update error: %v", err)
		return nil, errors.Wrap(err, "db.Collection.FindOneAndUpdate")
	}

	return &updatedCart, nil
}

// AddItem adds an item to the cart
func (r *cartMongoRepository) AddItem(ctx context.Context, userID string, item models.CartItem) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.AddItem")
	defer span.Finish()

	// First, try to get existing cart
	existingCart, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if existingCart == nil {
		// Create new cart with this item
		newCart := &models.Cart{
			UserID: userID,
			Items:  []models.CartItem{item},
		}
		return r.Create(ctx, newCart)
	}

	// Check if item already exists in cart
	itemExists := false
	for i, existingItem := range existingCart.Items {
		if existingItem.ProductID == item.ProductID {
			existingCart.Items[i].Quantity += item.Quantity
			itemExists = true
			break
		}
	}

	if !itemExists {
		existingCart.Items = append(existingCart.Items, item)
	}

	return r.Update(ctx, existingCart)
}

// UpdateItemQuantity updates the quantity of an item in the cart
func (r *cartMongoRepository) UpdateItemQuantity(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.UpdateItemQuantity")
	defer span.Finish()

	existingCart, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if existingCart == nil {
		return nil, errors.New("cart not found")
	}

	if quantity <= 0 {
		return r.RemoveItem(ctx, userID, productID)
	}

	itemFound := false
	for i, item := range existingCart.Items {
		if item.ProductID == productID {
			existingCart.Items[i].Quantity = quantity
			itemFound = true
			break
		}
	}

	if !itemFound {
		return nil, errors.New("item not found in cart")
	}

	return r.Update(ctx, existingCart)
}

// RemoveItem removes an item from the cart
func (r *cartMongoRepository) RemoveItem(ctx context.Context, userID string, productID primitive.ObjectID) (*models.Cart, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.RemoveItem")
	defer span.Finish()

	existingCart, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if existingCart == nil {
		return nil, errors.New("cart not found")
	}

	// Remove the item
	newItems := make([]models.CartItem, 0)
	for _, item := range existingCart.Items {
		if item.ProductID != productID {
			newItems = append(newItems, item)
		}
	}
	existingCart.Items = newItems

	return r.Update(ctx, existingCart)
}

// Clear removes all items from the cart
func (r *cartMongoRepository) Clear(ctx context.Context, userID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.Clear")
	defer span.Finish()

	_, err := r.db.Collection(cartsCollection).UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{
			"$set": bson.M{
				"items":     []models.CartItem{},
				"total":     0,
				"itemCount": 0,
				"updatedAt": time.Now(),
			},
		},
	)

	if err != nil {
		r.log.Errorf("Clear error: %v", err)
		return errors.Wrap(err, "db.Collection.UpdateOne")
	}

	return nil
}

// Delete deletes a cart
func (r *cartMongoRepository) Delete(ctx context.Context, userID string) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cartMongoRepository.Delete")
	defer span.Finish()

	_, err := r.db.Collection(cartsCollection).DeleteOne(ctx, bson.M{"userId": userID})
	if err != nil {
		r.log.Errorf("Delete error: %v", err)
		return errors.Wrap(err, "db.Collection.DeleteOne")
	}

	return nil
}
