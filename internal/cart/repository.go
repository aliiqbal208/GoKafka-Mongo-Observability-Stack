package cart

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aliiqbal208/GoKafka-Mongo-Observability-Stack/internal/models"
)

// MongoRepository defines the cart MongoDB repository interface
type MongoRepository interface {
	// GetByUserID retrieves a cart by user ID
	GetByUserID(ctx context.Context, userID string) (*models.Cart, error)

	// Create creates a new cart
	Create(ctx context.Context, cart *models.Cart) (*models.Cart, error)

	// Update updates an existing cart
	Update(ctx context.Context, cart *models.Cart) (*models.Cart, error)

	// AddItem adds an item to the cart
	AddItem(ctx context.Context, userID string, item models.CartItem) (*models.Cart, error)

	// UpdateItemQuantity updates the quantity of an item in the cart
	UpdateItemQuantity(ctx context.Context, userID string, productID primitive.ObjectID, quantity int64) (*models.Cart, error)

	// RemoveItem removes an item from the cart
	RemoveItem(ctx context.Context, userID string, productID primitive.ObjectID) (*models.Cart, error)

	// Clear removes all items from the cart
	Clear(ctx context.Context, userID string) error

	// Delete deletes a cart
	Delete(ctx context.Context, userID string) error
}

// RedisRepository defines the cart Redis cache repository interface
type RedisRepository interface {
	// GetByUserID retrieves a cart from cache
	GetByUserID(ctx context.Context, userID string) (*models.Cart, error)

	// SetCart caches a cart
	SetCart(ctx context.Context, cart *models.Cart) error

	// DeleteCart removes a cart from cache
	DeleteCart(ctx context.Context, userID string) error
}
